package transport

import "sync"

type ChanResponse struct {
	err    error
	ch     chan OperationResponse
	close  func() error
	closed bool

	cor OperationResponse
	m   sync.Mutex
}

func NewChanResponse(onClose func() error) *ChanResponse {
	return &ChanResponse{
		ch:    make(chan OperationResponse),
		close: onClose,
	}
}

func (r *ChanResponse) Next() bool {
	if r.err != nil {
		return false
	}

	or, ok := <-r.ch
	r.cor = or
	return ok
}

func (r *ChanResponse) Get() OperationResponse {
	return r.cor
}

func (r *ChanResponse) Close() {
	if r.close != nil {
		r.err = r.close()
	}
}

func (r *ChanResponse) CloseCh() {
	if r.closed {
		return
	}

	r.m.Lock()
	close(r.ch)
	r.closed = true
	r.m.Unlock()
}

func (r *ChanResponse) Err() error {
	return r.err
}

func (r *ChanResponse) Send(op OperationResponse) {
	r.m.Lock()
	if !r.closed {
		r.ch <- op
	}
	r.m.Unlock()
}
