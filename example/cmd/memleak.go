package main

import (
	"context"
	"example"
	"example/graph"
	"example/graph/generated"
	_ "expvar" // Register the expvar handlers
	"fmt"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	htransport "github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/websocket"
	"github.com/infiotinc/gqlgenc/client"
	"github.com/infiotinc/gqlgenc/client/transport"
	"net/http"
	_ "net/http/pprof" // Register the pprof handlers
	"runtime"
	"time"
)

func round(cli *client.Client) {
	var v interface{}
	err := cli.Query(context.Background(), "", example.RoomQuery, map[string]interface{}{"name": "test"}, &v)
	if err != nil {
		panic(err)
	}

	res, err := cli.Subscription(context.Background(), "", example.MessagesSub, nil)
	if err != nil {
		panic(err)
	}
	defer res.Close()

	for res.Next() {
		_ = res.Get()
	}

	if res.Err() != nil {
		panic(res.Err())
	}
}

func main() {
	srv := handler.New(generated.NewExecutableSchema(generated.Config{
		Resolvers: &graph.Resolver{},
	}))

	srv.AddTransport(htransport.POST{})
	srv.AddTransport(htransport.Websocket{
		KeepAlivePingInterval: 500 * time.Millisecond,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.Use(extension.Introspection{})

	http.Handle("/playground", playground.Handler("Playground", "/"))
	http.Handle("/", srv)

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()

	cli := &client.Client{
		Transport: &transport.Http{
			URL:            "http://localhost:8080",
		},
	}

	fmt.Println("Starting queries")

	for i := 0; i < 100000; i++ {
		if i % 1000 == 0 {
			fmt.Println(i)
		}

		round(cli)
	}

	fmt.Println("Running GC")

	runtime.GC()

	fmt.Println("Done")

	select {}
}
