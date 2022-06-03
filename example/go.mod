module example

replace github.com/infiotinc/gqlgenc => ../

go 1.15

require (
	github.com/99designs/gqlgen v0.17.9
	github.com/gorilla/websocket v1.5.0
	github.com/infiotinc/gqlgenc v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.1
	github.com/urfave/cli/v2 v2.8.1
	github.com/vektah/gqlparser/v2 v2.4.4
	golang.org/x/tools v0.1.10
	nhooyr.io/websocket v1.8.7
)
