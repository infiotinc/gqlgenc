package client

import (
	"context"
	"fmt"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

type Client struct {
	Transport Transport
}

func (c *Client) do(ctx context.Context, query string, variables map[string]interface{}) (Response, error) {
	q, errs := parser.ParseQuery(&ast.Source{Input: query})
	if errs != nil {
		return nil, errs
	}

	op := q.Operations[0]

	return c.Transport.Request(Request{
		Context:       ctx,
		OperationType: op.Operation,
		OperationName: op.Name,
		Query:         query,
		Variables:     variables,
	})
}

func (c *Client) Query(ctx context.Context, query string, variables map[string]interface{}, t interface{}) error {
	res, err := c.do(ctx, query, variables)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		res.Close()
	}()

	ok := res.Next()
	if !ok {
		return fmt.Errorf("no response")
	}

	opres := res.Get()
	err = opres.UnmarshalData(t)
	if err != nil {
		return err
	}

	return res.Err()
}

func (c *Client) Subscription(ctx context.Context, query string, variables map[string]interface{}) (Response, error) {
	res, err := c.do(ctx, query, variables)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		res.Close()
	}()

	return res, nil
}
