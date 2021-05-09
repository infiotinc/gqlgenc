package example

import (
	"context"
	"example/graph"
	"example/graph/generated"
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	htransport "github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/infiotinc/gqlgenc/client"
	"github.com/infiotinc/gqlgenc/client/transport"
	"github.com/stretchr/testify/assert"
	"log"
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

func wstr(ctx context.Context, u string) *transport.Ws {
	return cwstr(
		ctx,
		u,
		nil,
		time.Second, // We want immediate connection or fail
	)
}

func cwstr(ctx context.Context, u string, newWebsocketConn transport.WebsocketConnProvider, retryTimeout time.Duration) *transport.Ws {
	_ = os.Setenv("WS_LOG", "1")

	if strings.HasPrefix(u, "http") {
		u = "ws" + strings.TrimPrefix(u, "http")
	}

	tr := &transport.Ws{
		Context:               ctx,
		URL:                   u,
		RetryTimeout:          retryTimeout,
		WebsocketConnProvider: newWebsocketConn,
	}
	errCh := tr.Start()
	go func() {
		for err := range errCh {
			log.Println("Ws Transport error: ", err)
		}
	}()

	return tr
}

func httptr(ctx context.Context, u string) *transport.Http {
	tr := &transport.Http{
		URL: u,
	}

	return tr
}

func clifactory(ctx context.Context, trf func(server *httptest.Server) (transport.Transport, func())) (*client.Client, func()) {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{},
	}))

	srv.AddTransport(htransport.POST{})
	srv.AddTransport(htransport.Websocket{
		KeepAlivePingInterval: 400 * time.Millisecond,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		InitFunc: func(ctx context.Context, initPayload htransport.InitPayload) (context.Context, error) {
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
	err := cli.Query(ctx, "", roomQuery, nil, &opres)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "test", opres.Room.Name)
}

func runAssertSub(t *testing.T, ctx context.Context, cli *client.Client) {
	res, err := cli.Subscription(ctx, "", messagesSub, nil)
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
