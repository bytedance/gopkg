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
	"sync"
	"testing"
	"time"
)

func BenchmarkBreaker(b *testing.B) {
	op := Options{
		ShouldTrip: ConsecutiveTripFunc(1000),
	}
	cb, _ := newBreaker(op)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.IsAllowed()
		cb.Succeed()
	}
}

func BenchmarkBreakerParallel(b *testing.B) {
	op := Options{
		ShouldTrip: ConsecutiveTripFunc(1000),
	}
	cb, _ := newBreaker(op)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.IsAllowed()
			cb.Succeed()
		}
	})
}

func BenchmarkBreakerParallel2Cores(b *testing.B) {
	op := Options{
		ShouldTrip: ConsecutiveTripFunc(1000),
	}
	cb, _ := newBreaker(op)
	b.SetParallelism(2)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cb.IsAllowed()
			cb.Succeed()
		}
	})
}

func TestBreakerConsecutiveTrip(t *testing.T) {
	cooling := time.Millisecond * 50
	retry := time.Millisecond * 20

	op := Options{
		CoolingTimeout: cooling,
		DetectTimeout:  retry,
		ShouldTrip:     ConsecutiveTripFunc(1000),
	}
	cb, _ := newBreaker(op)

	for i := 0; i < 999; i++ {
		assert(t, cb.IsAllowed())
		cb.Fail()
	}

	assert(t, cb.IsAllowed())
	cb.Fail()
	assert(t, !cb.IsAllowed())

	time.Sleep(cooling)
	assert(t, cb.IsAllowed())
	cb.Fail()
	assert(t, !cb.IsAllowed())

	time.Sleep(cooling)
	assert(t, cb.IsAllowed())
	cb.Succeed()

	assert(t, !cb.IsAllowed())
	assert(t, cb.State() == HalfOpen)

	time.Sleep(retry)
	assert(t, cb.IsAllowed())
	cb.Succeed()
	assert(t, cb.State() == Closed)

	for i := 0; i < 100; i++ {
		assert(t, cb.IsAllowed())
		cb.Timeout()
	}

	assert(t, cb.IsAllowed())
	cb.Succeed()

	for i := 0; i < 1000; i++ {
		assert(t, cb.IsAllowed())
		cb.Timeout()
	}

	assert(t, !cb.IsAllowed())
}

func TestBreakerThresholdTrip(t *testing.T) {
	cooling := time.Millisecond * 50
	retry := time.Millisecond * 20

	op := Options{
		CoolingTimeout: cooling,
		DetectTimeout:  retry,
		ShouldTrip:     ThresholdTripFunc(1000),
	}

	cb, _ := newBreaker(op)

	for i := 0; i < 999; i++ {
		assert(t, cb.IsAllowed())
		cb.Timeout()
	}

	for i := 0; i < 1000; i++ {
		assert(t, cb.IsAllowed())
		cb.Succeed()
	}

	assert(t, cb.IsAllowed())
	cb.Fail()

	assert(t, !cb.IsAllowed())
}

func TestBreakerInstanceTrip(t *testing.T) {
	op := Options{
		ShouldTrip: ConsecutiveTripFuncV2(0.5, 1000, 3*time.Second, 50, 500),
	}

	cb, _ := newBreaker(op)

	for i := 0; i < 100; i++ {
		assert(t, cb.IsAllowed())
		cb.Timeout()
	}
	time.Sleep(3 * time.Second)
	assert(t, cb.IsAllowed())
	cb.Timeout()

	for i := 0; i < 100; i++ {
		assert(t, !cb.IsAllowed())
	}
}

func TestBreakerRateTrip(t *testing.T) {
	cooling := time.Millisecond * 50
	retry := time.Millisecond * 20

	op := Options{
		CoolingTimeout: cooling,
		DetectTimeout:  retry,
		ShouldTrip:     RateTripFunc(.5, 1000),
	}

	cb, _ := newBreaker(op)

	for i := 0; i < 499; i++ {
		assert(t, cb.IsAllowed())
		cb.Timeout()
	}

	for i := 0; i < 500; i++ {
		assert(t, cb.IsAllowed())
		cb.Succeed()
	}

	assert(t, cb.IsAllowed())
	cb.Fail()
	assert(t, !cb.IsAllowed())
	assert(t, cb.metricer.ErrorRate() == .5)

	time.Sleep(cooling)
	assert(t, cb.IsAllowed())
	cb.Succeed()

	assert(t, cb.State() == HalfOpen)
}

func TestBreakerRateTrip2(t *testing.T) {
	cooling := time.Millisecond
	retry := time.Millisecond / 10

	op := Options{
		CoolingTimeout: cooling,
		DetectTimeout:  retry,
		ShouldTrip:     RateTripFunc(.5, 1000),
	}

	cb, _ := newBreaker(op)

	for i := 0; i < 1000000; i++ {
		if cb.IsAllowed() == false {
			time.Sleep(retry)
			continue
		}

		r := rand.Intn(100)
		if r < 60 { // 60% fail
			cb.Fail()
		} else {
			cb.Succeed()
		}
	}

	if cb.metricer.Samples() > 1000 && cb.metricer.ErrorRate() >= .5 {
		assert(t, cb.State() != Closed)

	} else {
		assert(t, cb.State() == Closed)
	}
}

func TestBreakerReset(t *testing.T) {
	cooling := time.Millisecond
	retry := time.Millisecond / 10

	op := Options{
		CoolingTimeout: cooling,
		DetectTimeout:  retry,
		ShouldTrip:     RateTripFunc(.5, 1000),
	}

	cb, _ := newBreaker(op)

	for i := 0; i < 1000; i++ {
		assert(t, cb.IsAllowed())
		cb.Timeout()
	}

	assert(t, cb.metricer.ErrorRate() == 1)
	assert(t, cb.metricer.Samples() == 1000)
	assert(t, cb.metricer.Timeouts() == 1000)
	assert(t, !cb.IsAllowed())

	cb.Reset()

	assert(t, cb.metricer.ErrorRate() == 0)
	assert(t, cb.metricer.Samples() == 0)
	assert(t, cb.metricer.Timeouts() == 0)
	assert(t, cb.IsAllowed())
}

func TestBreakerConcurrent(t *testing.T) {
	cooling := time.Millisecond * 1000
	retry := time.Millisecond * 500
	opt := Options{
		CoolingTimeout: cooling,
		DetectTimeout:  retry,
		ShouldTrip:     RateTripFunc(0.5, 2),
	}
	var w sync.WaitGroup
	for i := 0; i < 10; i++ {
		w.Add(1)
		go func() {
			defer w.Done()
			b, _ := newBreaker(opt)
			// close -> open
			for i := 0; i < 2; i++ {
				b.Fail()
			}
			if b.State() != Open {
				t.Errorf("want open state but got %s", b.State())
			}

			// CoolingTimeout
			time.Sleep(cooling)
			var wg sync.WaitGroup
			b.IsAllowed()
			for i := 0; i < 49; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					assert(t, !b.IsAllowed())
				}()
			}
			wg.Wait()
		}()
	}
	w.Wait()
}
