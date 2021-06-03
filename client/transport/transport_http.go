package transport

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type HttpRequestOption func(req *http.Request)

type Http struct {
	URL string
	// Client defaults to http.DefaultClient
	Client         *http.Client
	RequestOptions []HttpRequestOption
}

func (h *Http) request(gqlreq Request) (*OperationResponse, error) {
	if h.Client == nil {
		h.Client = http.DefaultClient
	}

	bodyb, err := json.Marshal(NewOperationRequestFromRequest(gqlreq))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(gqlreq.Context, "POST", h.URL, bytes.NewReader(bodyb))
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

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var opres OperationResponse
	err = json.Unmarshal(data, &opres)
	if err != nil {
		return nil, err
	}

	return &opres, nil
}

func (h *Http) Request(req Request) Response {
	opres, err := h.request(req)
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewSingleResponse(*opres)
}
