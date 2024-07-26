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

package circuitbreaker

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkPerPBuckets(b *testing.B) {
	bk := newPerPBucket()
	bk.Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bk.Fail()
		bk.Timeout()
		bk.Succeed()
	}
}

// TestPerPMetricer1 tests basic functions
func TestPerPMetricer1(t *testing.T) {
	m := newPerPWindow()

	// no data
	deepEqual(t, m.ErrorRate(), float64(0))
	deepEqual(t, m.ConseErrors(), int64(0))
	deepEqual(t, m.Successes(), int64(0))
	deepEqual(t, m.Failures(), int64(0))
	deepEqual(t, m.Timeouts(), int64(0))

	m.Fail()
	m.Timeout()
	m.Fail()
	m.Timeout()
	deepEqual(t, m.Failures(), int64(2))
	deepEqual(t, m.Timeouts(), int64(2))
	deepEqual(t, m.ErrorRate(), float64(1))
	deepEqual(t, m.ConseErrors(), int64(4))
	deepEqual(t, m.Successes(), int64(0))

	m.Succeed()
	deepEqual(t, m.ErrorRate(), float64(.8))
	deepEqual(t, m.ConseErrors(), int64(0))
	deepEqual(t, m.Successes(), int64(1))
	deepEqual(t, m.Failures(), int64(2))
	deepEqual(t, m.Timeouts(), int64(2))

	s, f, tm := m.Counts()
	deepEqual(t, s, int64(1))
	deepEqual(t, f, int64(2))
	deepEqual(t, tm, int64(2))

	tot := 1000000
	for i := 0; i < tot; i++ {
		t := rand.Intn(3)
		if t == 0 { // fail
			m.Fail()
		} else if t == 1 { // timeout
			m.Timeout()
		} else { // succeed
			m.Succeed()
		}
	}

	rate := m.ErrorRate()
	assert(t, rate > .6 && rate < .7)
	s = m.Successes()
	assert(t, s > int64(tot/3-1000) && s < int64(tot/3+1000))
	f = m.Failures()
	assert(t, f > int64(tot/3-1000) && f < int64(tot/3+1000))
	ts := m.Timeouts()
	assert(t, ts > int64(tot/3-1000) && ts < int64(tot/3+1000))
}

// TestPerPMetricer2 tests functions about time
func TestPerPMetricer2(t *testing.T) {
	p, _ := NewPanel(nil, Options{BucketTime: time.Millisecond * 10, BucketNums: 100})
	b := p.(*panel).getBreaker("test")
	m := b.metricer
	expire := time.Millisecond * 10 * 100

	m.Succeed()
	deepEqual(t, m.Successes(), int64(1))
	deepEqual(t, m.Failures(), int64(0))
	deepEqual(t, m.Timeouts(), int64(0))

	time.Sleep(expire + time.Millisecond*10)
	deepEqual(t, m.Successes(), int64(0))
	deepEqual(t, m.Failures(), int64(0))
	deepEqual(t, m.Timeouts(), int64(0))

	for i := 0; i < 10; i++ {
		m.Fail()
	}
	deepEqual(t, m.Successes(), int64(0))
	deepEqual(t, m.Failures(), int64(10))
	deepEqual(t, m.Timeouts(), int64(0))
	deepEqual(t, m.ConseErrors(), int64(10))

	time.Sleep(expire / 2)
	for i := 0; i < 100; i++ {
		m.Fail()
	}
	deepEqual(t, m.Successes(), int64(0))
	deepEqual(t, m.Failures(), int64(110))
	deepEqual(t, m.Timeouts(), int64(0))
	deepEqual(t, m.ConseErrors(), int64(110))

	time.Sleep(expire / 2)
	deepEqual(t, m.Successes(), int64(0))
	deepEqual(t, m.Failures(), int64(100))
	deepEqual(t, m.Timeouts(), int64(0))
	deepEqual(t, m.ConseErrors(), int64(110))

	time.Sleep(expire / 2)
	deepEqual(t, m.Successes(), int64(0))
	deepEqual(t, m.Failures(), int64(0))
	deepEqual(t, m.Timeouts(), int64(0))
	deepEqual(t, m.ConseErrors(), int64(110))

	m.Succeed()
	deepEqual(t, m.Successes(), int64(1))
	deepEqual(t, m.Failures(), int64(0))
	deepEqual(t, m.Timeouts(), int64(0))
	deepEqual(t, m.ConseErrors(), int64(0))
}

func BenchmarkPerPWindow(b *testing.B) {
	m := newPerPWindow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Succeed()
	}
}

func BenchmarkPerPWindowParallel(b *testing.B) {
	m := newPerPWindow()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Succeed()
		}
	})
}

func BenchmarkPerPWindowParallel2Cores(b *testing.B) {
	b.SetParallelism(2)
	m := newPerPWindow()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Succeed()
		}
	})
}
