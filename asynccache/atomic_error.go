// Copyright 2021 ByteDance Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
