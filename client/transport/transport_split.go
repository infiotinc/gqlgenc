package transport

type Func func(Request) (Response, error)

func (f Func) Request(o Request) (Response, error) {
	return f(o)
}

func Split(f func(Request) (Transport, error)) Transport {
	return Func(func(req Request) (Response, error) {
		tr, err := f(req)
		if err != nil {
			return nil, err
		}

		return tr.Request(req)
	})
}

// SplitSubscription routes subscription to subtr, and other type of queries to othertr
func SplitSubscription(subtr, othertr Transport) Transport {
	return Split(func(req Request) (Transport, error) {
		if req.Operation == Subscription {
			return subtr, nil
		}

		return othertr, nil
	})
}
