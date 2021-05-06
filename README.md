# gqlgenc

> **Note**: ⚠️ This is a WIP, backward-compatibility cannot be guaranteed yet, use at your own risk

> gqlgenc is a fully featured go gql client, powered by codegen 

## Why yet another go GQL client ?

| Repo                                        | Codegen | Websocket Subscription |
|---------------------------------------------|---------|------------------------|
| https://github.com/shurcooL/graphql         | ❌      | ❌                      |
| https://github.com/Yamashou/gqlgenc         | ✅      | ❌                      |
| https://github.com/hasura/go-graphql-client | ❌      | ✅                      |
| ✨https://github.com/infiotinc/gqlgenc✨     | ✅      | ✅                      |

## GQL Client

### Transports

- http: Transports GQL queries over http
- ws: Transports GQL queries over websocket
- split: Can be used to have a single client use multiple transports depending on the type of query (`query`, `mutation` over http and `subscription` over ws)

### Quickstart

Quickstart with a client with http & ws transports:

```go
package main

import (
    "context"
    "github.com/infiotinc/gqlgenc/client"
    "github.com/vektah/gqlparser/v2/ast"
    "time"
)

func main() {
    ctx := context.Background()

    wstr := &client.WsTransport{
        Context: ctx,
        URL:     "ws://example.org/graphql",
        Timeout: time.Minute,
        RetryTimeout: time.Second,
    }
    wstr.Start()
    defer wstr.Close()

    httptr := &client.HttpTransport{
        URL:    "http://example.org/graphql",
    }

    tr := client.SplitTransport(func(req client.Request) (client.Transport, error) {
        if req.OperationType == ast.Subscription {
            return wstr, nil
        }

        return httptr, nil
    })

    cli := &client.Client {
        Transport: tr,
    }
}
```

### Query/Mutation

```go
var res struct {
    Room string `json:"room"`
}
err := cli.Query(ctx, "query { room }", nil, &res)
if err != nil {
    panic(err)
}
```

### Subscription

```go
sub, err := cli.Subscription(ctx, "subscription { newRoom }", nil)
if err != nil {
    panic(err)
}

for sub.Next() {
    msg := sub.Get()
    
    if len(msg.Errors) > 0 {
        // Do something with them
    }
    
    var res struct {
        Room string `json:"newRoom"`
    }
    err := msg.UnmarshalData(&res)
    if err != nil {
        // Do something with that
    }
}

if err := sub.Err(); err != nil {
    panic(err)
}
```

## GQL Client Codegen

Create a `.gqlgenc.yml` at the root of your module:

```yaml
model:
  package: graph
  filename: ./graph/gen_models.go
client:
  package: graph
  filename: ./graph/gen_client.go
models:
  Int:
    model: github.com/99designs/gqlgen/graphql.Int64
  DateTime:
    model: github.com/99designs/gqlgen/graphql.Time
# The schema can be fetched from files or through introspection
schema:
  - schema.graphqls
endpoint:
  url: https://api.annict.com/graphql # Where do you want to send your request?
  headers:　# If you need header for getting introspection query, set it
    Authorization: "Bearer ${ANNICT_KEY}" # support environment variables
query:
  - query.graphql

```

Fill your `query.graphql` with queries:
```graphql
query GetRoom {
    room(name: "secret room") {
        name
    }
}
```

Run `go run github.com/infiotinc/gqlgenc`

Enjoy:
```go
// Create codegen client
gql := &graph.Client{
    Client: cli,
}

gql.GetRoom(...)
```

