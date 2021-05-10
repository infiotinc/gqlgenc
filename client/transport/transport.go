package transport

import (
	"context"
	"encoding/json"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type Operation string

const (
	Query        Operation = "query"
	Mutation     Operation = "mutation"
	Subscription Operation = "subscription"
)

type OperationRequest struct {
	Query         string                 `json:"query"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

type OperationResponse struct {
	Errors gqlerror.List   `json:"errors,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

func (r OperationResponse) UnmarshalData(t interface{}) error {
	if r.Data == nil {
		return nil
	}

	return json.Unmarshal(r.Data, t)
}

type Request struct {
	Context       context.Context
	Operation     Operation
	OperationName string

	Query     string
	Variables map[string]interface{}
}

type Response interface {
	Next() bool
	Get() OperationResponse
	Close()
	Err() error
}

type Transport interface {
	Request(req Request) (Response, error)
}
