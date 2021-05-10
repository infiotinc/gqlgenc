package transport

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type HttpRequestOption func(req *http.Request)

type Http struct {
	URL            string
	// Client defaults to http.DefaultClient
	Client         *http.Client
	RequestOptions []HttpRequestOption
}

func (h *Http) Request(o Request) (Response, error) {
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

	return NewSingleResponse(opres), nil
}
