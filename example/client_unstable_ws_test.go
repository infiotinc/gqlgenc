package example

import (
	"context"
	"fmt"
	"github.com/infiotinc/gqlgenc/client"
	"github.com/infiotinc/gqlgenc/client/transport"
	"math/rand"
	"net/http/httptest"
	"nhooyr.io/websocket"
	"testing"
	"time"
)

type unstableWebsocketConn struct {
	ctx    context.Context
	wsconn *transport.WebsocketHandler
}

func (u *unstableWebsocketConn) dropConn() {
	fmt.Println("## DROP CONN")
	_ = u.wsconn.Conn.Close(websocket.StatusProtocolError, "conn drop")
}

func (u *unstableWebsocketConn) mayErr() error {
	if rand.Float64() < 0.4 { // simulate a slow one
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		return nil
	}

	return nil
}

func (u *unstableWebsocketConn) ReadJSON(v interface{}) error {
	if rand.Float64() < 0.4 { // simulate context deadline early
		return context.DeadlineExceeded
	}

	if err := u.mayErr(); err != nil {
		return err
	}

	return u.wsconn.ReadJSON(v)
}

func (u *unstableWebsocketConn) WriteJSON(v interface{}) error {
	if err := u.mayErr(); err != nil {
		return err
	}

	return u.wsconn.WriteJSON(v)
}

func (u *unstableWebsocketConn) Close() error {
	return u.wsconn.Close()
}

func (u *unstableWebsocketConn) SetReadLimit(limit int64) {
	u.wsconn.SetReadLimit(limit)
}

func newUnstableConn(ctx context.Context, URL string) (transport.WebsocketConn, error) {
	rand.Seed(time.Now().UTC().UnixNano())

	wsconn, err := transport.DefaultWebsocketConnProvider(500*time.Millisecond)(ctx, URL)
	if err != nil {
		return nil, err
	}

	return &unstableWebsocketConn{
		ctx:    ctx,
		wsconn: wsconn.(*transport.WebsocketHandler),
	}, nil
}

func unstablewscli(ctx context.Context, newWebsocketConn transport.WebsocketConnProvider) (*client.Client, func()) {
	return clifactory(ctx, func(ts *httptest.Server) (transport.Transport, func()) {
		tr := cwstr(ctx, ts.URL, newWebsocketConn, 0)

		return tr, func() {
			tr.Close()
		}
	})
}

func TestRawWSUnstableQuery(t *testing.T) {
	ctx := context.Background()

	cli, teardown := unstablewscli(ctx, newUnstableConn)
	defer teardown()
	tr := cli.Transport.(*transport.Ws)

	time.Sleep(2*time.Second)

	for i := 0; i < 5; i++ {
		fmt.Println("> Attempt", i)
		tr.GetConn().(*unstableWebsocketConn).dropConn()

		time.Sleep(2*time.Second)

		runAssertQuery(t, ctx, cli)
	}
}

func TestRawWSUnstableSubscription(t *testing.T) {
	ctx := context.Background()

	cli, teardown := unstablewscli(ctx, newUnstableConn)
	defer teardown()
	tr := cli.Transport.(*transport.Ws)

	time.Sleep(2*time.Second)

	for i := 0; i < 5; i++ {
		fmt.Println("> Attempt", i)
		tr.GetConn().(*unstableWebsocketConn).dropConn()

		time.Sleep(2*time.Second)

		runAssertSub(t, ctx, cli)
	}
}
