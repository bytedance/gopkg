package channel

import (
	"context"
	"sync/atomic"
	"time"
)

var (
	_ Cond = (*cond)(nil)
)

type CondOption func(c *cond)

func WithCondTimeout(timeout time.Duration) CondOption {
	return func(c *cond) {
		c.timeout = timeout
	}
}

type Cond interface {
	Signal() bool
	Broadcast() bool
	Wait(ctx context.Context) bool
}

func NewCond(opts ...CondOption) Cond {
	return new(cond)
}

type condSignal = chan struct{}

type cond struct {
	signal  atomic.Value
	timeout time.Duration
}

func (c *cond) Signal() bool {
	sv := c.signal.Load()
	if sv == nil {
		return false
	}
	signal := sv.(condSignal)
	select {
	case signal <- struct{}{}:
		return true
	default:
		return false
	}
}

func (c *cond) Broadcast() bool {
BROADCAST:
	sv := c.signal.Load()
	if sv == nil {
		return false
	}
	var signal condSignal = nil
	if !c.signal.CompareAndSwap(sv, signal) {
		goto BROADCAST
	}
	signal = sv.(condSignal)
	select {
	case <-signal:
		return false
	default:
		close(signal)
		return true
	}
}

func (c *cond) Wait(ctx context.Context) bool {
WAIT:
	sv := c.signal.Load()
	var signal condSignal
	if sv == nil {
		signal = make(condSignal)
		if !c.signal.CompareAndSwap(nil, signal) {
			goto WAIT
		}
	} else {
		signal = sv.(condSignal)
	}
	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}
	if ctx == nil || ctx.Done() == nil {
		<-signal
		return true
	}
	select {
	case <-signal:
		return true
	case <-ctx.Done():
		return false
	}
}
