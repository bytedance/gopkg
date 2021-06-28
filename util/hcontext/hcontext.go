package hcontext

import (
	"context"
	"time"
)

type valueOnlyContext struct{ context.Context }

func (valueOnlyContext) Deadline() (deadline time.Time, ok bool) { return }
func (valueOnlyContext) Done() <-chan struct{}                   { return nil }
func (valueOnlyContext) Err() error                              { return nil }

//WithNoDeadline 移除超时控制，保留value信息
func WithNoDeadline(ctx context.Context) context.Context {
	return valueOnlyContext{ctx}
}

//WithNoCancel 不受上层context cancel的影响，但是保留上层ctx的Deadline信息
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
