example-gqlgen:
	cd example && go run github.com/99designs/gqlgen

example-gqlgenc:
	cd example && go run github.com/infiotinc/gqlgenc

example-test:
	cd example && go test -v -count=1 ./...
