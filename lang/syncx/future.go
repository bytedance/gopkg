// Copyright 2024 ByteDance Inc.
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

package syncx

import (
	"sync/atomic"
)

const (
	stateFree uint64 = iota
	stateGray
	stateDone
)

type state struct {
	noCopy noCopy

	state uint64 // high 32 bits are state, low 32 bits are waiter count.
	sema  uint32

	val interface{}
	err error
}

type Promise struct {
	state state
}

type Future struct {
	state *state
}

func (s *state) set(val interface{}, err error) {
	for {
		state := atomic.LoadUint64(&s.state)
		if (state >> 32) > stateFree {
			panic("promise already satisfied")
		}
		if atomic.CompareAndSwapUint64(&s.state, state, state+(1<<32)) {
			s.val = val
			s.err = err
			state := atomic.AddUint64(&s.state, 1)
			for w := state & (1<<32 - 1); w > 0; w-- {
				runtime_Semrelease(&s.sema, false, 0)
			}
			return
		}
	}
}

func (s *state) get() (interface{}, error) {
	for {
		state := atomic.LoadUint64(&s.state)
		if (state >> 32) == stateDone {
			return s.val, s.err
		}
		if atomic.CompareAndSwapUint64(&s.state, state, state+1) {
			runtime_Semacquire(&s.sema)
			return s.val, s.err
		}
	}
}

func NewPromise() *Promise {
	return &Promise{}
}

func (p *Promise) Set(val interface{}, err error) {
	p.state.set(val, err)
}

func (p *Promise) Future() *Future {
	return &Future{state: &p.state}
}

func (f *Future) Get() (interface{}, error) {
	return f.state.get()
}
