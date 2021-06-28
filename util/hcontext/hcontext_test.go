package hcontext

import (
	"context"
	"testing"
	"time"

	"gopkg.in/go-playground/assert.v1"
)

type ctxtype string

var k ctxtype = "k"

func TestWithNoDeadline(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		stub func(*testing.T, context.Context)
	}{
		{
			name: "value正常传递",
			args: args{
				context.WithValue(context.TODO(), k, "world"),
			},
			stub: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, ctx.Value(k).(string), "world")
			},
		},
		{
			name: "nil Deadline()",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
					_ = cancel
					return ctx
				}(),
			},
			stub: func(t *testing.T, ctx context.Context) {
				deadline, ok := ctx.Deadline()
				assert.Equal(t, deadline.IsZero(), true)
				assert.Equal(t, ok, false)

			},
		},
		{
			name: "nil Done()",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithTimeout(context.TODO(), time.Microsecond)
					_ = cancel
					return ctx
				}(),
			},
			stub: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, ctx.Done(), nil)
			},
		},
		{
			name: "nil Err()",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
					cancel()
					return ctx
				}(),
			},
			stub: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, ctx.Err(), nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithNoDeadline(tt.args.ctx)
			tt.stub(t, got)
		})
	}
}

func TestWithNoCancel(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		stub func(*testing.T, context.Context)
	}{
		{
			name: "value正常传递",
			args: args{
				context.WithValue(context.TODO(), k, "world"),
			},
			stub: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, ctx.Value(k).(string), "world")
			},
		},

		{
			name: "new Deadline()",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithDeadline(context.TODO(), time.Date(2022, 1, 1, 1, 1, 1, 1, time.Local))
					_ = cancel
					return ctx
				}(),
			},
			stub: func(t *testing.T, ctx context.Context) {
				deadline, ok := ctx.Deadline()
				assert.Equal(t, ok, true)
				assert.Equal(t, deadline, time.Date(2022, 1, 1, 1, 1, 1, 1, time.Local))
			},
		},

		{
			name: "cancel withtimeout context",
			args: args{
				ctx: func() context.Context {
					now := time.Now()
					ctx := context.WithValue(context.Background(), "t", now)
					ctx, cancel := context.WithDeadline(ctx, now.Add(1000*time.Microsecond))
					go func() {
						time.Sleep(100 * time.Microsecond)
						cancel()
					}()
					return ctx
				}(),
			},

			stub: func(t *testing.T, ctx context.Context) {
				<-ctx.Done()
				duration := time.Since(ctx.Value("t").(time.Time))
				if ctx.Err() == context.Canceled {
					t.Error("context canceled")
				}
				if duration < 200*time.Microsecond {
					t.Error("timeout cancel ")
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithNoCancel(tt.args.ctx)
			tt.stub(t, got)
		})
	}
}
