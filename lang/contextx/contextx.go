package contextx

import (
	"context"
	"time"
)

type valueOnlyContext struct{ context.Context }

func (valueOnlyContext) Deadline() (deadline time.Time, ok bool) { return }
func (valueOnlyContext) Done() <-chan struct{}                   { return nil }
func (valueOnlyContext) Err() error                              { return nil }

//RemoveDeadline only remove deadline value
func RemoveDeadline(ctx context.Context) context.Context {
	return valueOnlyContext{ctx}
}

//RemoveCancel only remove cancel value
func RemoveCancel(ctx context.Context) context.Context {
	deadline, ok := ctx.Deadline()
	if ok {
		var cancel context.CancelFunc
		ctx = RemoveDeadline(ctx)
		ctx, cancel = context.WithDeadline(ctx, deadline)
		_ = cancel
		return ctx
	}
	return RemoveDeadline(ctx)
}
