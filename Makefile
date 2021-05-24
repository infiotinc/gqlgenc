example-gqlgen:
	cd example && go run github.com/99designs/gqlgen

example-gqlgenc:
	cd example && go run github.com/infiotinc/gqlgenc

example-test:
	cd example && go test -v -count=1 ./...

example-run-memleak:
	cd example && go run ./cmd/memleak.go

tag:
	git tag -a ${TAG} -m ${TAG}
	git push origin ${TAG}
