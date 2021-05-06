module example

replace github.com/infiotinc/gqlgenc => ../

go 1.16

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/gorilla/websocket v1.4.2
	github.com/infiotinc/gqlgenc v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli/v2 v2.1.1
	github.com/vektah/gqlparser/v2 v2.1.0
	golang.org/x/tools v0.0.0-20200827163409-021d7c6f1ec3
)
