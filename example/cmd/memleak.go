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
	"sync"
	"sync/atomic"
	"time"
)

func round(cli *client.Client) {
	var v interface{}
	err := cli.Query(context.Background(), "", example.RoomQuery, map[string]interface{}{"name": "test"}, &v)
	if err != nil {
		fmt.Println("ERROR QUERY: ", err)
	}

	res, err := cli.Subscription(context.Background(), "", example.MessagesSub, nil)
	if err != nil {
		fmt.Println("ERROR SUB: ", err)
	}
	defer res.Close()

	for res.Next() {
		_ = res.Get()
	}

	if res.Err() != nil {
		fmt.Println("ERROR SUB RES: ", res.Err())
	}
}

func main() {
	{
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

		httptr := &transport.Http{
			URL: "http://localhost:8080",
		}
		wstr := &transport.Ws{
			URL:                   "ws://localhost:8080",
			WebsocketConnProvider: transport.DefaultWebsocketConnProvider(time.Second),
		}
		wstr.Start(context.Background())

		tr := transport.SplitSubscription(wstr, httptr)

		cli := &client.Client{
			Transport: tr,
		}

		fmt.Println("Starting queries")

		var wg sync.WaitGroup
		ch := make(chan struct{}, 5)
		var di int64
		for i := 0; i < 100_000; i++ {
			wg.Add(1)

			ch <- struct{}{}
			go func() {
				round(cli)
				di := atomic.AddInt64(&di, 1)
				if di%1000 == 0 {
					fmt.Println(di)
				}
				<-ch
				wg.Done()
			}()
		}

		wg.Wait()
		wstr.Close()
	}

	time.Sleep(2 * time.Second)

	fmt.Println("Running GC")
	runtime.GC()

	fmt.Println("Done")

	select {}
}
