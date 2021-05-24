package transport

import "sync"

type SingleResponse struct {
	or OperationResponse

	calledNext bool
	dm sync.Mutex
	dc chan struct{}
}

func NewSingleResponse(or OperationResponse) *SingleResponse {
	return &SingleResponse{or: or}
}

func (r *SingleResponse) Next() bool {
	defer func() {
		r.calledNext = true
	}()

	return !r.calledNext
}

func (r *SingleResponse) Get() OperationResponse {
	return r.or
}

func (r *SingleResponse) Close() {}

func (r *SingleResponse) Done() <-chan struct{} {
	r.dm.Lock()
	if r.dc == nil {
		r.dc = make(chan  struct{})
		close(r.dc)
	}
	r.dm.Unlock()

	return r.dc
}

func (r *SingleResponse) Err() error {
	return nil
}
