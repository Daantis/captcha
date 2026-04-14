package server

import "sync/atomic"

type incomingStreamCloser interface {
	closeIncomingStream() error
	closed() bool
}

type closer struct {
	isClosed atomic.Bool
}

func (c *closer) closeIncomingStream() error {
	c.isClosed.Store(true)
	return nil
}

func (c *closer) closed() bool {
	return c.isClosed.Load()
}

func newCloser() *closer {
	return &closer{}
}
