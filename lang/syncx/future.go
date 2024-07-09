package syncx

import (
	"sync"
	"sync/atomic"
)

type state struct {
	noCopy noCopy

	m    sync.Mutex
	c    *sync.Cond
	done int32
	val  interface{}
	err  error
}

type Promise struct {
	state *state
}

type Future struct {
	state *state
}

func newState() *state {
	s := &state{}
	s.c = sync.NewCond(&s.m)
	return s
}

func (s *state) set(val interface{}, err error) {
	s.m.Lock()
	defer s.m.Unlock()
	if atomic.LoadInt32(&s.done) == 1 {
		panic("promise already satisfied")
	}
	s.val = val
	s.err = err
	atomic.StoreInt32(&s.done, 1)
	s.c.Broadcast()
}

func (s *state) get() (interface{}, error) {
	if atomic.LoadInt32(&s.done) == 1 {
		return s.val, s.err
	}
	s.m.Lock()
	defer s.m.Unlock()
	for atomic.LoadInt32(&s.done) != 1 {
		s.c.Wait()
	}
	return s.val, s.err
}

func NewPromise() *Promise {
	return &Promise{state: newState()}
}

func (p *Promise) Set(val interface{}, err error) {
	p.state.set(val, err)
}

func (p *Promise) Future() *Future {
	return &Future{state: p.state}
}

func (f *Future) Get() (interface{}, error) {
	return f.state.get()
}
