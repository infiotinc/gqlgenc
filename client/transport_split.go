package client

type FuncTransport func(Request) (Response, error)

func (f FuncTransport) Request(o Request) (Response, error) {
	return f(o)
}

type funcSplitTransport func(Request) (Transport, error)

func SplitTransport(f funcSplitTransport) Transport {
	return FuncTransport(func(req Request) (Response, error) {
		tr, err := f(req)
		if err != nil {
			return nil, err
		}

		return tr.Request(req)
	})
}

// SplitSubscription routes subscription to subtr, and other type of queries to othertr
func SplitSubscription(subtr, othertr Transport) Transport {
	return SplitTransport(func(req Request) (Transport, error) {
		if req.Operation == Subscription {
			return subtr, nil
		}

		return othertr, nil
	})
}
