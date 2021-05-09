package example

import (
	"context"
	"github.com/infiotinc/gqlgenc/client"
	"github.com/infiotinc/gqlgenc/client/transport"
	"net/http/httptest"
	"testing"
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
