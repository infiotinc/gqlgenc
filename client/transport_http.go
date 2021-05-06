package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type httpResponse struct {
	or OperationResponse

	calledNext bool
}

func (r *httpResponse) Next() bool {
	defer func() {
		r.calledNext = true
	}()

	return !r.calledNext
}

func (r httpResponse) Get() OperationResponse {
	return r.or
}

func (r *httpResponse) Close() {}

func (r httpResponse) Err() error {
	return nil
}

type HttpRequestOption func(req *http.Request)

type HttpTransport struct {
	Client         *http.Client
	URL            string
	RequestOptions []HttpRequestOption
}

func (h *HttpTransport) Request(o Request) (Response, error) {
	if h.Client == nil {
		h.Client = http.DefaultClient
	}

	body := OperationRequest{
		Query:         o.Query,
		OperationName: o.OperationName,
		Variables:     o.Variables,
	}

	bodyb, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(o.Context, "POST", h.URL, bytes.NewReader(bodyb))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	for _, ro := range h.RequestOptions {
		ro(req)
	}

	res, err := h.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var opres OperationResponse
	err = json.Unmarshal(data, &opres)
	if err != nil {
		return nil, err
	}

	return &httpResponse{
		or: opres,
	}, nil
}
