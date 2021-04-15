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
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func BenchmarkBuckets(b *testing.B) {
	bk := bucket{}
	bk.Reset()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bk.Fail()
		bk.Timeout()
		bk.Succeed()
	}
}

// TestMetricser1 tests basic functions
func TestMetricser1(t *testing.T) {
	m := newWindow()

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
	assert(t, (rate > .6 && rate < .7))
	s = m.Successes()
	assert(t, (s > int64(tot/3-1000) && s < int64(tot/3+1000)))
	f = m.Failures()
	assert(t, (f > int64(tot/3-1000) && f < int64(tot/3+1000)))
	ts := m.Timeouts()
	assert(t, (ts > int64(tot/3-1000) && ts < int64(tot/3+1000)))
}

// TestMetricser2 tests functions about time
func TestMetricser2(t *testing.T) {
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

func TestMetricser3(t *testing.T) {
	p, _ := NewPanel(nil, Options{BucketTime: time.Millisecond * 10, BucketNums: 100})
	b := p.(*panel).getBreaker("test")
	m := b.metricer
	worker := func(qps int, m metricer) {
		interval := time.Second / time.Duration(qps)
		for range time.Tick(interval) {
			m.Succeed()
		}
	}

	for i := 0; i < 10; i++ {
		go worker(1000, m)
	}

	time.Sleep(time.Second)

	cnt := 0
	for range time.Tick(time.Second) {
		s := m.Successes()
		fmt.Println(s)
		assert(t, (s < 11000 && s > 9000))

		cnt++
		if cnt == 3 {
			break
		}
	}
}

func TestMetricser4(t *testing.T) {
	m := newWindow()
	for i := 0; i < 10; i++ {
		r := rand.Intn(1000) + 1000
		for j := 0; j < r; j++ {
			m.Fail()
			m.Timeout()
			m.Succeed()
		}

		deepEqual(t, m.Failures(), int64(r))
		deepEqual(t, m.Timeouts(), int64(r))
		deepEqual(t, m.Successes(), int64(r))
		deepEqual(t, m.ConseErrors(), int64(0))
		rate := m.ErrorRate()
		assert(t, rate > .6 && rate < .7)

		m.Reset()
	}
}

func TestMetricser5(t *testing.T) {
	_, err := newWindowWithOptions(time.Millisecond, 99)
	assert(t, err != nil)

	p, _ := NewPanel(nil, Options{BucketTime: time.Millisecond, BucketNums: 100})
	b := p.(*panel).getBreaker("test")
	m := b.metricer
	expire := time.Millisecond * 101

	for i := 0; i < 105; i++ {
		m.Succeed()
		time.Sleep(time.Millisecond)
	}

	time.Sleep(expire)

	assert(t, m.Samples() == 0)
	assert(t, m.ErrorRate() == 0)
}

func TestMetricser6(t *testing.T) {
	_, err := newWindowWithOptions(time.Millisecond, 99)
	assert(t, err != nil)

	p, _ := NewPanel(nil, Options{BucketTime: time.Millisecond, BucketNums: 100})
	b := p.(*panel).getBreaker("test")
	m := b.metricer

	for i := 0; i < 105; i++ {
		m.Fail()
		time.Sleep(time.Millisecond)
	}

	assert(t, m.ConseTime() > 0)
	assert(t, m.ConseErrors() > 0)
}

func BenchmarkWindow(b *testing.B) {
	m := newWindow()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Succeed()
	}
}

func BenchmarkWindowParallel(b *testing.B) {
	m := newWindow()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Succeed()
		}
	})
}

func BenchmarkWindowParallel2Cores(b *testing.B) {
	b.SetParallelism(2)
	m := newWindow()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Succeed()
		}
	})
}
