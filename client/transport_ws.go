package client

// Original work from https://github.com/hasura/go-graphql-client/blob/0806e5ec7/subscription.go

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"nhooyr.io/websocket"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type OperationMessageType string

const (
	// GQL_CONNECTION_INIT the Client sends this message after plain websocket connection to start the communication with the server
	GQL_CONNECTION_INIT OperationMessageType = "connection_init"
	// GQL_CONNECTION_ERROR The server may responses with this message to the GQL_CONNECTION_INIT from client, indicates the server rejected the connection.
	GQL_CONNECTION_ERROR OperationMessageType = "conn_err"
	// GQL_START Client sends this message to execute GraphQL operation
	GQL_START OperationMessageType = "start"
	// GQL_STOP Client sends this message in order to stop a running GraphQL operation execution (for example: unsubscribe)
	GQL_STOP OperationMessageType = "stop"
	// GQL_ERROR Server sends this message upon a failing operation, before the GraphQL execution, usually due to GraphQL validation errors (resolver errors are part of GQL_DATA message, and will be added as errors array)
	GQL_ERROR OperationMessageType = "error"
	// GQL_DATA The server sends this message to transfter the GraphQL execution result from the server to the client, this message is a response for GQL_START message.
	GQL_DATA OperationMessageType = "data"
	// GQL_COMPLETE Server sends this message to indicate that a GraphQL operation is done, and no more data will arrive for the specific operation.
	GQL_COMPLETE OperationMessageType = "complete"
	// GQL_CONNECTION_KEEP_ALIVE Server message that should be sent right after each GQL_CONNECTION_ACK processed and then periodically to keep the client connection alive.
	// The client starts to consider the keep alive message only upon the first received keep alive message from the server.
	GQL_CONNECTION_KEEP_ALIVE OperationMessageType = "ka"
	// GQL_CONNECTION_ACK The server may responses with this message to the GQL_CONNECTION_INIT from client, indicates the server accepted the connection. May optionally include a payload.
	GQL_CONNECTION_ACK OperationMessageType = "connection_ack"
	// GQL_CONNECTION_TERMINATE the Client sends this message to terminate the connection.
	GQL_CONNECTION_TERMINATE OperationMessageType = "connection_terminate"

	// GQL_UNKNOWN is an Unknown operation type, for logging only
	GQL_UNKNOWN OperationMessageType = "unknown"
	// GQL_INTERNAL is the Internal status, for logging only
	GQL_INTERNAL OperationMessageType = "internal"
)

type WebsocketConn interface {
	ReadJSON(v interface{}) error
	WriteJSON(v interface{}) error
	Close() error
	// SetReadLimit sets the maximum size in bytes for a message read from the peer. If a
	// message exceeds the limit, the connection sends a close message to the peer
	// and returns ErrReadLimit to the application.
	SetReadLimit(limit int64)
}

type OperationMessage struct {
	ID      string               `json:"id,omitempty"`
	Type    OperationMessageType `json:"type"`
	Payload json.RawMessage      `json:"payload,omitempty"`
}

func (msg OperationMessage) String() string {
	return fmt.Sprintf("%v %v %s", msg.ID, msg.Type, msg.Payload)
}

type ConnOptions struct {
	Context context.Context
	URL     string
	Timeout time.Duration
}

type wsResponse struct {
	Request

	close func() error

	ch      chan OperationResponse
	cor     OperationResponse
	err     error
	started bool
}

func (r *wsResponse) Next() bool {
	if r.err != nil {
		return false
	}

	or, ok := <-r.ch
	r.cor = or
	return ok
}

func (r wsResponse) Get() OperationResponse {
	return r.cor
}

func (r *wsResponse) Close() {
	r.err = r.close()
}

func (r wsResponse) Err() error {
	return r.err
}

type WsTransport struct {
	ConnOptions
	ConnectionParams interface{}
	NewWebsocketConn func(o ConnOptions) (WebsocketConn, error)
	RetryTimeout     time.Duration

	conn        WebsocketConn
	operations  map[string]*wsResponse
	operationsm sync.Mutex
	started     bool
	isRunning   bool
	context     context.Context
	cancel      context.CancelFunc

	o     sync.Once
	wsLog bool
}

func (t *WsTransport) initStruct() {
	t.o.Do(func() {
		if t.operations == nil {
			t.operations = map[string]*wsResponse{}
		}

		t.wsLog, _ = strconv.ParseBool(os.Getenv("WS_LOG"))
	})
}

func (t *WsTransport) Start() {
	go func() {
		err := t.Run()
		if err != nil {
			panic(err)
		}
	}()
}

func (t *WsTransport) setIsRunning(value bool) {
	t.printLog(GQL_INTERNAL, "TRY ISRUNNING", value)
	t.operationsm.Lock()
	t.isRunning = value
	t.printLog(GQL_INTERNAL, "ISRUNNING SET", value)
	t.operationsm.Unlock()
}

func (t *WsTransport) Run() error {
	t.printLog(GQL_INTERNAL, "RUN")

	t.initStruct()

	if err := t.init(); err != nil {
		return fmt.Errorf("retry timeout. exiting...")
	}

	t.printLog(GQL_INTERNAL, "INIT DONE")

	t.setIsRunning(true)

	t.printLog(GQL_INTERNAL, "FULLY RUNNING")

	for t.isRunning {
		select {
		case <-t.context.Done():
			return nil
		default:
			var message OperationMessage
			if err := t.conn.ReadJSON(&message); err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					continue // Is expected as part of conn.ReadJSON
				}

				t.printLog(GQL_INTERNAL, "READ ERR", err)

				if err == io.EOF || strings.Contains(err.Error(), "EOF") {
					return t.Reset()
				}
				closeStatus := websocket.CloseStatus(err)
				if closeStatus == websocket.StatusNormalClosure {
					// close event from websocket client, exiting...
					return nil
				}
				if closeStatus != -1 {
					return t.Reset()
				}
				continue
			}

			switch message.Type {
			case GQL_ERROR:
				t.printLog(GQL_ERROR, message)
				fallthrough
			case GQL_DATA:
				t.printLog(GQL_DATA, message)

				id := message.ID
				sub, ok := t.operations[id]
				if !ok {
					continue
				}

				var out OperationResponse
				err := json.Unmarshal(message.Payload, &out)
				if err != nil {
					out.Error = err
				}
				sub.ch <- out
			case GQL_CONNECTION_ERROR:
				t.printLog(GQL_CONNECTION_ERROR, message)
			case GQL_COMPLETE:
				t.printLog(GQL_COMPLETE, message)
				_ = t.unsubscribe(message.ID)
			case GQL_CONNECTION_KEEP_ALIVE:
				t.printLog(GQL_CONNECTION_KEEP_ALIVE, message)
			case GQL_CONNECTION_ACK:
				t.printLog(GQL_CONNECTION_ACK, message)
				for k, v := range t.operations {
					if err := t.startSubscription(k, v); err != nil {
						_ = t.unsubscribe(k)
						return err
					}
				}
			default:
				t.printLog(GQL_UNKNOWN, message)
			}
		}
	}

	// if the running status is false, stop retrying
	if !t.isRunning {
		return nil
	}

	return t.Reset()
}

func (t *WsTransport) Reset() error {
	t.printLog(GQL_INTERNAL, "RESET")

	if !t.isRunning {
		return nil
	}

	for id, op := range t.operations {
		_ = t.stopSubscription(id)
		op.started = false
	}

	if t.conn != nil {
		_ = t.terminate()
		_ = t.conn.Close()
		t.conn = nil
	}
	t.cancel()

	return t.Run()
}

func (t *WsTransport) Close() error {
	t.setIsRunning(false)
	for id := range t.operations {
		if err := t.unsubscribe(id); err != nil {
			t.cancel()
			return err
		}
	}

	var err error

	if t.conn != nil {
		_ = t.terminate()
		err = t.conn.Close()
		t.conn = nil
	}
	t.cancel()

	return err
}

func (t *WsTransport) startSubscription(id string, res *wsResponse) error {
	if res == nil || res.started {
		return nil
	}

	t.printLog(GQL_INTERNAL, "START SUB")

	in := OperationRequest{
		Query:         res.Query,
		OperationName: res.OperationName,
		Variables:     res.Variables,
	}

	payload, err := json.Marshal(in)
	if err != nil {
		return err
	}

	msg := OperationMessage{
		ID:      id,
		Type:    GQL_START,
		Payload: payload,
	}

	t.printLog(GQL_START, msg)
	if err := t.conn.WriteJSON(msg); err != nil {
		return err
	}

	go func() {
		<-res.Context.Done()
		_ = t.unsubscribe(id)
	}()

	res.started = true

	return nil
}

func (t *WsTransport) stopSubscription(id string) error {
	if t.conn == nil {
		return nil
	}

	msg := OperationMessage{
		ID:   id,
		Type: GQL_STOP,
	}

	t.printLog(GQL_STOP, msg)
	return t.conn.WriteJSON(msg)
}

func (t *WsTransport) unsubscribe(id string) error {
	t.operationsm.Lock()
	res, ok := t.operations[id]
	if !ok {
		return fmt.Errorf("subscription id %s doesn't not exist", id)
	}

	err := t.stopSubscription(id)

	close(res.ch)

	delete(t.operations, id)
	t.operationsm.Unlock()
	return err
}

func (t *WsTransport) terminate() error {
	if t.conn != nil {
		// send terminate message to the server
		msg := OperationMessage{
			Type: GQL_CONNECTION_TERMINATE,
		}

		t.printLog(GQL_CONNECTION_TERMINATE, msg)
		return t.conn.WriteJSON(msg)
	}

	return nil
}

func (t *WsTransport) Request(req Request) (Response, error) {
	t.printLog(GQL_INTERNAL, "REQ")

	t.initStruct()

	id := fmt.Sprintf("%p-%v", &req, time.Now().Nanosecond())

	res := &wsResponse{
		Request: req,
		close: func() error {
			return t.unsubscribe(id)
		},
		ch: make(chan OperationResponse),
	}

	if t.isRunning {
		err := t.startSubscription(id, res)
		if err != nil {
			return nil, err
		}
	}

	t.printLog(GQL_INTERNAL, "ADD TO OPS")
	t.operationsm.Lock()
	t.operations[id] = res
	t.operationsm.Unlock()

	return res, nil
}

func (t *WsTransport) sendConnectionInit() error {
	var bParams []byte = nil
	if t.ConnectionParams != nil {
		var err error
		bParams, err = json.Marshal(t.ConnectionParams)
		if err != nil {
			return err
		}
	}

	msg := OperationMessage{
		Type:    GQL_CONNECTION_INIT,
		Payload: bParams,
	}

	t.printLog(GQL_CONNECTION_INIT, msg)
	return t.conn.WriteJSON(msg)
}

func (t *WsTransport) init() error {
	t.printLog(GQL_INTERNAL, "INIT")

	start := time.Now()
	ctx, cancel := context.WithCancel(context.Background())
	t.context = ctx
	t.cancel = cancel

	if t.NewWebsocketConn == nil {
		t.NewWebsocketConn = NewWebsocketConn()
	}

	for {
		var err error
		var conn WebsocketConn
		// allow custom websocket client
		if t.conn == nil {
			conn, err = t.NewWebsocketConn(t.ConnOptions)
			if err == nil {
				t.conn = conn
			}
		}

		if err == nil {
			//t.conn.SetReadLimit(t.readLimit)
			err = t.sendConnectionInit()
		}
		if err == nil {
			return nil
		}

		if time.Now().After(start.Add(t.RetryTimeout)) {
			return err
		}

		t.printLog(GQL_INTERNAL, err.Error()+". retry in second....")
		time.Sleep(time.Second)
	}
}

func (t *WsTransport) printLog(typ OperationMessageType, rest ...interface{}) {
	if t.wsLog {
		fmt.Printf("# %-20v: ", typ)
		fmt.Println(rest...)
	}
}
