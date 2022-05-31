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

package lscq

import (
	"sync"
	"testing"
	"unsafe"

	"github.com/bytedance/gopkg/collection/skipset"
	"github.com/bytedance/gopkg/lang/fastrand"
)

func TestBoundedQueue(t *testing.T) {
	q := newUint64SCQ()
	s := skipset.NewUint64()

	// Dequeue empty queue.
	val, ok := q.Dequeue()
	if ok {
		t.Fatal(val)
	}

	// Single goroutine correctness.
	for i := 0; i < scqsize; i++ {
		if !q.Enqueue(uint64(i)) {
			t.Fatal(i)
		}
		s.Add(uint64(i))
	}

	if q.Enqueue(20) { // queue is full
		t.Fatal()
	}

	s.Range(func(value uint64) bool {
		if val, ok := q.Dequeue(); !ok || val != value {
			t.Fatal(val, ok, value)
		}
		return true
	})

	// Dequeue empty queue after previous loop.
	val, ok = q.Dequeue()
	if ok {
		t.Fatal(val)
	}

	// ---------- MULTIPLE TEST BEGIN ----------.
	for j := 0; j < 10; j++ {
		s = skipset.NewUint64()

		// Dequeue empty queue.
		val, ok = q.Dequeue()
		if ok {
			t.Fatal(val)
		}

		// Single goroutine correctness.
		for i := 0; i < scqsize; i++ {
			if !q.Enqueue(uint64(i)) {
				t.Fatal()
			}
			s.Add(uint64(i))
		}

		if q.Enqueue(20) { // queue is full
			t.Fatal()
		}

		s.Range(func(value uint64) bool {
			if val, ok := q.Dequeue(); !ok || val != value {
				t.Fatal(val, ok, value)
			}
			return true
		})

		// Dequeue empty queue after previous loop.
		val, ok = q.Dequeue()
		if ok {
			t.Fatal(val)
		}
	}
	// ---------- MULTIPLE TEST END ----------.

	// MPMC correctness.
	var wg sync.WaitGroup
	s1 := skipset.NewUint64()
	s2 := skipset.NewUint64()
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func() {
			if fastrand.Uint32n(2) == 0 {
				r := fastrand.Uint64()
				if q.Enqueue(r) {
					s1.Add(r)
				}
			} else {
				val, ok := q.Dequeue()
				if ok {
					s2.Add(uint64(val))
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	for {
		val, ok := q.Dequeue()
		if !ok {
			break
		}
		s2.Add(uint64(val))
	}

	s1.Range(func(value uint64) bool {
		if !s2.Contains(value) {
			t.Fatal(value)
		}
		return true
	})

	if s1.Len() != s2.Len() {
		t.Fatal("invalid")
	}
}

func TestUnboundedQueue(t *testing.T) {
	// MPMC correctness.
	q := NewUint64()
	var wg sync.WaitGroup
	s1 := skipset.NewUint64()
	s2 := skipset.NewUint64()
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func() {
			if fastrand.Uint32n(2) == 0 {
				r := fastrand.Uint64()
				if !s1.Add(r) || !q.Enqueue(r) {
					panic("invalid")
				}
			} else {
				val, ok := q.Dequeue()
				if ok {
					s2.Add(uint64(val))
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	for {
		val, ok := q.Dequeue()
		if !ok {
			break
		}
		s2.Add(uint64(val))
	}

	s1.Range(func(value uint64) bool {
		if !s2.Contains(value) {
			t.Fatal(value)
		}
		return true
	})

	if s1.Len() != s2.Len() {
		t.Fatal("invalid")
	}
}

type foo struct {
	val int
}

func TestPointerQueue(t *testing.T) {
	q := NewPointer()

	for i := 0; i < 10; i++ {
		q.Enqueue(unsafe.Pointer(&foo{val: i}))
	}

	for i := 0; i < 10; i++ {
		if p, ok := q.Dequeue(); !ok || (*foo)(p).val != i {
			t.Fatal("got:", (*foo)(p).val, ok, "expect:", i, true)
		}
	}
}
