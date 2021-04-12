package asynccache

import "sync/atomic"

// Error is an atomic type-safe wrapper around Value for errors
type Error struct{ v atomic.Value }

// errorHolder is non-nil holder for error object.
// atomic.Value panics on saving nil object, so err object needs to be
// wrapped with valid object first.
type errorHolder struct{ err error }

// Load atomically loads the wrapped error
func (e *Error) Load() error {
	v := e.v.Load()
	if v == nil {
		return nil
	}

	eh := v.(errorHolder)
	return eh.err
}

// Store atomically stores error.
// NOTE: a holder object is allocated on each Store call.
func (e *Error) Store(err error) {
	e.v.Store(errorHolder{err: err})
}
