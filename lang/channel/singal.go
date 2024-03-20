package channel

import (
	"context"
	"time"
)

var (
	_ Signal = (*sigal)(nil)
)

type Signal interface {
	Signal()
	Wait(ctx context.Context) bool
}

type SignalOption func(c *sigal)

func WithSignalTimeout(timeout time.Duration) SignalOption {
	return func(s *sigal) {
		s.timeout = timeout
	}
}

func NewSignal(opts ...SignalOption) Signal {
	sg := new(sigal)
	for _, opt := range opts {
		opt(sg)
	}
	sg.trigger = make(chan struct{})
	return sg
}

type sigal struct {
	trigger chan struct{}
	timeout time.Duration
}

func (s *sigal) Signal() {
	select {
	case <-s.trigger:
	default:
		close(s.trigger)
	}
}

func (s *sigal) Wait(ctx context.Context) bool {
	if s.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}
	if ctx == nil || ctx.Done() == nil {
		<-s.trigger
		return true
	}
	select {
	case <-s.trigger:
		return true
	case <-ctx.Done():
		return false
	}
}
