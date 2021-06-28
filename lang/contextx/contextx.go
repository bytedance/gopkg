package contextx

import (
	"context"
	"time"
)

type valueOnlyContext struct{ context.Context }

func (valueOnlyContext) Deadline() (deadline time.Time, ok bool) { return }
func (valueOnlyContext) Done() <-chan struct{}                   { return nil }
func (valueOnlyContext) Err() error                              { return nil }

//WithNoDeadline only remove deadline value
func WithNoDeadline(ctx context.Context) context.Context {
	return valueOnlyContext{ctx}
}

//WithNoCancel only remove cancel value
func WithNoCancel(ctx context.Context) context.Context {
	deadline, ok := ctx.Deadline()
	if ok {
		var cancel context.CancelFunc
		ctx = WithNoDeadline(ctx)
		ctx, cancel = context.WithDeadline(ctx, deadline)
		_ = cancel
		return ctx
	}
	return WithNoDeadline(ctx)
}
