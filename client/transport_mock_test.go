package client

import (
	"context"
	"fmt"
	"github.com/infiotinc/gqlgenc/client/transport"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSingleResponse(t *testing.T) {
	_ = os.Setenv("GQLGENC_WS_LOG", "1")

	cli := &Client{
		Transport: transport.Mock{
			"query": func(req transport.Request) (transport.Response, error) {
				return transport.NewSingleResponse(transport.NewMockOperationResponse("hey", nil)), nil
			},
		},
	}

	var res string
	err := cli.Query(context.Background(), "", "query", nil, &res)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "hey", res)
}

func TestChanResponse(t *testing.T) {
	_ = os.Setenv("GQLGENC_WS_LOG", "1")

	cli := &Client{
		Transport: transport.Mock{
			"query": func(req transport.Request) (transport.Response, error) {
				res := transport.NewChanResponse(nil)

				go func() {
					for i := 0; i < 3; i++ {
						res.Send(transport.NewMockOperationResponse(fmt.Sprintf("hey%v", i), nil))
					}
					res.CloseCh()
				}()

				return res, nil
			},
		},
	}

	sub, err := cli.Subscription(context.Background(), "", "query", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	msgs := make([]string, 0)

	for sub.Next() {
		res := sub.Get()

		var data string
		err := res.UnmarshalData(&data)
		if err != nil {
			t.Fatal(err)
		}

		msgs = append(msgs, data)
	}

	if err := sub.Err(); err != nil {
		t.Fatal(err)
	}

	assert.Len(t, msgs, 3)
}
