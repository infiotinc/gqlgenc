package example

import (
	"context"
	"example/graph"
	"example/graph/generated"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/infiotinc/gqlgenc/client"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const roomQuery = `
query query {
	room(name: "test") {
		name
	}
}
`

type RoomQueryResponse struct {
	Room struct {
		Name string `json:"name"`
	} `json:"room"`
}

const messagesSub = `
subscription query{
	messageAdded(roomName: "test") {
		id
	}
}
`

type MessagesSubResponse struct {
	MessageAdded struct {
		ID string `json:"id"`
	} `json:"messageAdded"`
}

func wstr(ctx context.Context, u string) *client.WsTransport {
	_ = os.Setenv("WS_LOG", "1")

	if strings.HasPrefix(u, "http") {
		u = "ws" + strings.TrimPrefix(u, "http")
	}

	tr := &client.WsTransport{
		ConnOptions: client.ConnOptions{
			Context: ctx,
			URL:     u,
			Timeout: time.Minute,
		},
		RetryTimeout: time.Second,
	}
	tr.Start()

	return tr
}

func httptr(ctx context.Context, u string) *client.HttpTransport {
	tr := &client.HttpTransport{
		Client: http.DefaultClient,
		URL:    u,
	}

	return tr
}

func clifactory(ctx context.Context, trf func(server *httptest.Server) (client.Transport, func())) (*client.Client, func()) {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{},
	}))

	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		InitFunc: func(ctx context.Context, initPayload transport.InitPayload) (context.Context, error) {
			fmt.Println("WS Init")

			return ctx, nil
		},
	})
	srv.Use(extension.Introspection{})

	httpsrv := http.NewServeMux()
	httpsrv.Handle("/playground", playground.Handler("Playground", "/"))
	httpsrv.Handle("/", srv)

	ts := httptest.NewServer(httpsrv)

	fmt.Println("TS URL: ", ts.URL)

	tr, trteardown := trf(ts)

	return &client.Client{
			Transport: tr,
		}, func() {
			if trteardown != nil {
				trteardown()
			}

			if ts != nil {
				ts.Close()
			}
		}
}

func runAssertQuery(t *testing.T, ctx context.Context, cli *client.Client) {
	var opres RoomQueryResponse
	err := cli.Query(ctx, roomQuery, nil, &opres)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "test", opres.Room.Name)
}

func runAssertSub(t *testing.T, ctx context.Context, cli *client.Client) {
	res, err := cli.Subscription(ctx, messagesSub, nil)
	if err != nil {
		t.Fatal(err)
	}

	ids := make([]string, 0)

	for res.Next() {
		op := res.Get()

		var opres MessagesSubResponse
		err := op.UnmarshalData(&opres)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, opres.MessageAdded.ID)
	}

	if res.Err() != nil {
		t.Fatal(err)
	}
	assert.Len(t, ids, 3)
}
