package webssh

import (
	"context"
	"net"
)

type inbound struct {
	ctx context.Context
	ln  net.Listener
	fwd forward
}

func newInbound(ctx context.Context, addr string, r forward) (*inbound, error) {
	listen, err := net.Listen("tcp", addr)
	if nil != err {
		return nil, err
	}
	return &inbound{ctx: ctx, ln: listen, fwd: r}, nil
}

func (in *inbound) startServe() {
	for {
		select {
		case <-in.ctx.Done():
			return
		default:
			c, err := in.ln.Accept()
			if nil != err {
				c.Close()
			} else {
				go in.fwd(c)
			}
		}
	}
}

func (in *inbound) Close() error {
	if nil != in {
		return in.ln.Close()
	}
	return nil
}
