// Package gopoolsession wraps gopool to support localsession automatic binding and unbinding.
// It keeps compatibility with the original gopool API, minimizing business code changes.
package gopoolsession

import (
	"context"

	"github.com/cloudwego/localsession"

	"github.com/bytedance/gopkg/util/gopool"
)

// Go submits an asynchronous task to the goroutine pool.
// It automatically gets the session from the current calling goroutine,
// binds it to the task goroutine, and unbinds it after the task completes.
// The signature is fully compatible with gopool.Go, so business code can switch with minimal changes.
func Go(task func()) {
	// Get the current session from the calling goroutine (no need for business to pass manually)
	currentSession, _ := localsession.CurSession()

	// Wrap the task with session lifecycle management (bind -> execute -> unbind)
	wrappedTask := wrapTaskWithAutoSession(currentSession, task)

	// Submit the wrapped task to the original gopool
	gopool.Go(wrappedTask)
}

// CtxGo submits an asynchronous task to the goroutine pool with context.
// It automatically gets the session from the current calling goroutine,
// binds it to the task goroutine, and unbinds it after the task completes.
// The signature is fully compatible with gopool.CtxGo, so business code can switch with minimal changes.
func CtxGo(ctx context.Context, task func()) {
	// Get the current session from the calling goroutine (no need for business to pass manually)
	currentSession, _ := localsession.CurSession()

	// Wrap the task with session lifecycle management (bind -> execute -> unbind)
	wrappedTask := wrapTaskWithAutoSession(currentSession, task)

	// Submit the wrapped task to the original gopool
	gopool.CtxGo(ctx, wrappedTask)
}

// wrapTaskWithAutoSession wraps the business task to implement the complete session lifecycle.
// It binds the session before task execution and unbinds it after task completion (even if panic occurs).
func wrapTaskWithAutoSession(sess localsession.Session, task func()) func() {
	return func() {
		// needUnbind marks whether to unbind the session after task execution
		// Only unbind when a valid session is bound to avoid meaningless operations
		needUnbind := false

		// Bind the session to the worker goroutine in the pool before task execution
		if sess != nil {
			localsession.BindSession(sess)
			needUnbind = true

			// Ensure the session is unbound after the task completes, regardless of success or failure
			// Use defer to guarantee execution even if the task panics
			defer func() {
				if needUnbind {
					localsession.UnbindSession()
				}
			}()
		}

		// Execute the original business task (the session is already bound to the current goroutine)
		if task != nil {
			task()
		}
	}
}
