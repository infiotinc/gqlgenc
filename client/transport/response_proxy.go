package transport

import (
	"sync"
)

type proxyTarget SendResponse

type proxyBind struct {
	proxy   *ProxyResponse
	res     Response
	onOpres func(response OperationResponse, send func())
}

func (pb *proxyBind) run() {
	res := pb.res
	defer pb.proxy.Unbind(res)

	go func() {
		select {
		case <-res.Done():
		case <-pb.proxy.Done():
			res.Close()
		}
	}()

	for res.Next() {
		opres := res.Get()

		pb.onOpres(opres, func() {
			if pb.proxy.Bound(res) {
				pb.proxy.Send(opres)
			}
		})

		if !pb.proxy.Bound(res) {
			break
		}
	}

	if pb.proxy.Bound(res) {
		if err := res.Err(); err != nil {
			pb.proxy.CloseWithError(err)
		}
	}
}

type ProxyResponse struct {
	proxyTarget
	binds []*proxyBind
	m     sync.RWMutex
}

func (p *ProxyResponse) Bound(res Response) bool {
	p.m.RLock()
	defer p.m.RUnlock()

	for _, b := range p.binds {
		if b.res == res {
			return true
		}
	}

	return false
}

func (p *ProxyResponse) Bind(res Response, onOpres func(response OperationResponse, send func())) {
	if onOpres == nil {
		onOpres = func(_ OperationResponse, send func()) {
			send()
		}
	}

	p.m.Lock()
	b := &proxyBind{
		proxy:   p,
		res:     res,
		onOpres: onOpres,
	}
	p.binds = append(p.binds, b)
	p.m.Unlock()

	go b.run()
}

func (p *ProxyResponse) Unbind(res Response) {
	p.m.Lock()
	pbc := len(p.binds)
	binds := make([]*proxyBind, 0)
	for _, b := range p.binds {
		if b.res == res {

		} else {
			binds = append(binds, b)
		}
	}
	p.binds = binds
	bc := len(p.binds)
	p.m.Unlock()

	if pbc != bc && bc == 0 {
		p.Close()
	}
}

func NewProxyResponse() *ProxyResponse {
	return &ProxyResponse{
		proxyTarget: NewChanResponse(nil),
	}
}
