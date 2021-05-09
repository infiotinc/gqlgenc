package transport

// Original work from https://github.com/hasura/go-graphql-client/blob/0806e5ec7/subscription.go

import (
	"context"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"time"
)

// WebsocketHandler is default websocket handler implementation using https://github.com/nhooyr/websocket
type WebsocketHandler struct {
	ctx     context.Context
	timeout time.Duration
	*websocket.Conn
}

func (wh *WebsocketHandler) WriteJSON(v interface{}) error {
	ctx, cancel := context.WithTimeout(wh.ctx, wh.timeout)
	defer cancel()

	return wsjson.Write(ctx, wh.Conn, v)
}

func (wh *WebsocketHandler) ReadJSON(v interface{}) error {
	ctx, cancel := context.WithTimeout(wh.ctx, wh.timeout)
	defer cancel()
	return wsjson.Read(ctx, wh.Conn, v)
}

func (wh *WebsocketHandler) Close() error {
	return wh.Conn.Close(websocket.StatusNormalClosure, "close websocket")
}

type WsDialOption func(o *websocket.DialOptions)

func DefaultWebsocketConnProvider(timeout time.Duration, optionfs ...WsDialOption) WebsocketConnProvider {
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

		return &WebsocketHandler{
			ctx:     ctx,
			Conn:    c,
			timeout: timeout,
		}, nil
	}
}
