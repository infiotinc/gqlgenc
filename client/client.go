package client

import (
	"context"
	"fmt"
	"github.com/vektah/gqlparser/v2/ast"
)

type Client struct {
	Transport Transport
}

type Operation ast.Operation

const (
	Query        Operation = "query"
	Mutation     Operation = "mutation"
	Subscription Operation = "subscription"
)

func (c *Client) do(ctx context.Context, operation Operation, operationName string, query string, variables map[string]interface{}, t interface{}) error {
	res, err := c.Transport.Request(Request{
		Context:       ctx,
		Operation:     operation,
		Query:         query,
		OperationName: operationName,
		Variables:     variables,
	})
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

	if err := res.Err(); err != nil {
		return err
	}

	opres := res.Get()
	err = opres.UnmarshalData(t)

	if len(opres.Errors) > 0 {
		return opres.Errors
	}

	return err
}

// Query runs a query
// operationName is optional
func (c *Client) Query(ctx context.Context, operationName string, query string, variables map[string]interface{}, t interface{}) error {
	return c.do(ctx, Query, operationName, query, variables, t)
}

// Mutation runs a mutation
// operationName is optional
func (c *Client) Mutation(ctx context.Context, operationName string, query string, variables map[string]interface{}, t interface{}) error {
	return c.do(ctx, Mutation, operationName, query, variables, t)
}

// Subscription starts a GQL subscription
// operationName is optional
func (c *Client) Subscription(ctx context.Context, operationName string, query string, variables map[string]interface{}) (Response, error) {
	res, err := c.Transport.Request(Request{
		Context:       ctx,
		Operation:     Subscription,
		Query:         query,
		OperationName: operationName,
		Variables:     variables,
	})
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		res.Close()
	}()

	return res, nil
}
