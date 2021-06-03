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
	Query         string                 `json:"query,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Extensions    map[string]interface{} `json:"extensions,omitempty"`
}

func NewOperationRequestFromRequest(req *Request) OperationRequest {
	return OperationRequest{
		Query:         req.Query,
		OperationName: req.OperationName,
		Variables:     req.Variables,
		Extensions:    req.Extensions.Map(),
	}
}

type OperationResponse struct {
	Data       json.RawMessage            `json:"data,omitempty"`
	Errors     gqlerror.List              `json:"errors,omitempty"`
	Extensions map[string]json.RawMessage `json:"extensions,omitempty"`
}

func (r OperationResponse) UnmarshalData(t interface{}) error {
	if r.Data == nil {
		return nil
	}

	return json.Unmarshal(r.Data, t)
}

func (r OperationResponse) UnmarshalExtension(name string, t interface{}) error {
	if r.Extensions == nil {
		return nil
	}

	ex, ok := r.Extensions[name]
	if !ok {
		return nil
	}

	return json.Unmarshal(ex, t)
}

type Extensions struct {
	m map[string]interface{}
}

func (e *Extensions) Get(k string) interface{} {
	if e.m == nil {
		return nil
	}

	return e.m[k]
}

func (e *Extensions) Set(k string, v interface{}) {
	if e.m == nil {
		e.m = make(map[string]interface{})
	}

	e.m[k] = v
}

func (e *Extensions) Map() map[string]interface{} {
	return e.m
}

func (e *Extensions) Has(key string) bool {
	if e.m == nil {
		return false
	}

	_, has := e.m[key]

	return has
}

type Request struct {
	Context   context.Context
	Operation Operation

	OperationName string
	Query         string
	Variables     map[string]interface{}
	Extensions    Extensions
}

type Transport interface {
	Request(req *Request) Response
}
