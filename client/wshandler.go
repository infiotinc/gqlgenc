package client

// Original work from https://github.com/hasura/go-graphql-client/blob/0806e5ec7/subscription.go

import (
	"context"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

// default websocket handler implementation using https://github.com/nhooyr/websocket
type websocketHandler struct {
	ctx     context.Context
	timeout time.Duration
	*websocket.Conn
}

func (wh *websocketHandler) WriteJSON(v interface{}) error {
	ctx, cancel := context.WithTimeout(wh.ctx, wh.timeout)
	defer cancel()

	return wsjson.Write(ctx, wh.Conn, v)
}

func (wh *websocketHandler) ReadJSON(v interface{}) error {
	ctx, cancel := context.WithTimeout(wh.ctx, wh.timeout)
	defer cancel()
	return wsjson.Read(ctx, wh.Conn, v)
}

func (wh *websocketHandler) Close() error {
	return wh.Conn.Close(websocket.StatusNormalClosure, "close websocket")
}

type WsDialOption func(o *websocket.DialOptions)

func NewWebsocketConn(timeout time.Duration, optionfs ...WsDialOption) NewWebsocketConnFunc {
	return func(ctx context.Context, URL string) (WebsocketConn, error) {
		options := &websocket.DialOptions{
			Subprotocols: []string{"graphql-ws"},
		}
		for _, f := range optionfs {
			f(options)
		}

		c, _, err := websocket.Dial(ctx, URL, options)
		if err != nil {
			return nil, err
		}

		return &websocketHandler{
			ctx:     ctx,
			Conn:    c,
			timeout: timeout,
		}, nil
	}
}
