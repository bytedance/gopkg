package lscq

import (
	"sync"
	"testing"

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
