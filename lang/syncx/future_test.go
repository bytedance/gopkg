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
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var errFoo = errors.New("foo")

func TestPromiseAndFuture(t *testing.T) {
	p := NewPromise()
	f := p.Future()
	p.Set(1, errFoo)
	val, err := f.Get()
	assert.Equal(t, val, 1)
	assert.Equal(t, err, errFoo)
}

func TestPromiseAndFutureConcurrency(t *testing.T) {
	n := runtime.NumCPU() - 1

	ch := make(chan struct{}, n)
	p := NewPromise()
	go func() {
		for i := 0; i < n; i++ {
			ch <- struct{}{}
		}
		time.Sleep(1 * time.Second)
		p.Set(1, errFoo)
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ch
			f := p.Future()
			val, err := f.Get()
			assert.Equal(t, val, 1)
			assert.Equal(t, err, errFoo)
		}()
	}
	wg.Wait()
}

func TestPromiseSetTwice(t *testing.T) {
	p := NewPromise()
	p.Set(1, nil)
	assert.Panics(t, func() {
		p.Set(1, nil)
	})
}

func Benchmark(b *testing.B) {
	b.Run("Promise", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			p := NewPromise()
			f := p.Future()
			go func() {
				p.Set(nil, nil)
			}()
			_, _ = f.Get()
		}
	})
	b.Run("WaitGroup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var val interface{}
			var err error
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				val, err = nil, nil
				wg.Done()
			}()
			wg.Wait()
			_, _ = val, err
		}
	})
	// channel does not support multi-consumers. This is used to compare the performance of a single consumer.
	b.Run("Channel", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var val interface{}
			var err error
			ch := make(chan struct{})
			go func() {
				val, err = nil, nil
				ch <- struct{}{}
			}()
			<-ch
			_, _ = val, err
		}
	})
}
