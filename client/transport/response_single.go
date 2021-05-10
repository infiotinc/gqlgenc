package transport

type SingleResponse struct {
	or OperationResponse

	calledNext bool
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

func (r SingleResponse) Get() OperationResponse {
	return r.or
}

func (r *SingleResponse) Close() {}

func (r SingleResponse) Err() error {
	return nil
}
