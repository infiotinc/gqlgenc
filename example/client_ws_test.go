package example

import (
	"context"
	"github.com/infiotinc/gqlgenc/client"
	"net/http/httptest"
	"testing"
)

func wscli(ctx context.Context) (*client.Client, func()) {
	return clifactory(ctx, func(ts *httptest.Server) (client.Transport, func()) {
		tr := wstr(ctx, ts.URL)

		return tr, func() {
			tr.Close()
		}
	})
}

func TestRawWSQuery(t *testing.T) {
	ctx := context.Background()

	cli, teardown := wscli(ctx)
	defer teardown()

	runAssertQuery(t, ctx, cli)
}

func TestRawWSSubscription(t *testing.T) {
	ctx := context.Background()

	cli, teardown := wscli(ctx)
	defer teardown()

	runAssertSub(t, ctx, cli)
}
