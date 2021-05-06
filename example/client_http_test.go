package example

import (
	"context"
	"github.com/infiotinc/gqlgenc/client"
	"net/http/httptest"
	"testing"
)

func httpcli(ctx context.Context) (*client.Client, func()) {
	return clifactory(ctx, func(ts *httptest.Server) (client.Transport, func()) {
		return httptr(ctx, ts.URL), nil
	})
}

func TestRawHttpQuery(t *testing.T) {
	ctx := context.Background()

	cli, teardown := httpcli(ctx)
	defer teardown()

	runAssertQuery(t, ctx, cli)
}
