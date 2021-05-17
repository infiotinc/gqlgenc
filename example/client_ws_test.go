package example

import (
	"context"
	"github.com/infiotinc/gqlgenc/client"
	"github.com/infiotinc/gqlgenc/client/transport"
	"net/http/httptest"
	"testing"
	"time"
)

func wscli(ctx context.Context, newWebsocketConn transport.WebsocketConnProvider) (*client.Client, func()) {
	return clifactory(ctx, func(ts *httptest.Server) (transport.Transport, func()) {
		tr := wstr(ctx, ts.URL)

		return tr, func() {
			tr.Close()
		}
	})
}

func TestRawWSQuery(t *testing.T) {
	ctx := context.Background()

	cli, teardown := wscli(ctx, nil)
	defer teardown()

	runAssertQuery(t, ctx, cli)
}

func TestRawWSSubscription(t *testing.T) {
	ctx := context.Background()

	cli, teardown := wscli(ctx, nil)
	defer teardown()

	runAssertSub(t, ctx, cli)
}

func TestWSCtxCancel1(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cli, teardown := wscli(ctx, nil)

	runAssertQuery(t, ctx, cli)

	teardown()
	time.Sleep(time.Second)
	cancel()
}

func TestWSCtxCancel2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	cli, teardown := wscli(ctx, nil)

	runAssertQuery(t, ctx, cli)

	cancel()
	time.Sleep(time.Second)
	teardown()
}
