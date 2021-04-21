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
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func testRateTripFuncFalse(rate float64, minSamples int64) TripFunc {
	return func(m Metricer) bool {
		samples := m.Samples()
		_ = samples >= minSamples && m.ErrorRate() >= rate
		return false
	}
}

func TestPanel(t *testing.T) {
	cooling := time.Millisecond * 10

	op := Options{
		CoolingTimeout: cooling,
		ShouldTrip:     ConsecutiveTripFunc(1000),
	}

	p, err := NewPanel(nil, op)
	assert(t, err == nil)
	if p == nil {
	}

	var counter int64
	var wg sync.WaitGroup
	worker := func() {
		for i := 0; i < 10; i++ {
			if p.IsAllowed("xxx") {
				atomic.AddInt64(&counter, 1)
				time.Sleep(time.Millisecond)
				atomic.AddInt64(&counter, -1)
				p.Succeed("xxx")
			}
			time.Sleep(time.Millisecond)
		}
		wg.Done()
	}

	checker := func() {
		for i := 0; i < 20; i++ {
			assert(t, atomic.LoadInt64(&counter) <= 20)
			time.Sleep(time.Millisecond)
		}
		wg.Done()
	}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker()
	}
	wg.Add(1)
	go checker()

	wg.Wait()
	assert(t, counter == 0)
}

func BenchmarkPanelClosed_IsAllowed(b *testing.B) {
	panel, err := NewPanel(nil, Options{})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		panel.IsAllowed(key)
	}
}

func BenchmarkPanelClosed_Succeed(b *testing.B) {
	panel, err := NewPanel(nil, Options{})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		panel.Succeed(key)
	}
}

func BenchmarkPanelClosed_Fail(b *testing.B) {
	panel, err := NewPanel(nil, Options{
		CoolingTimeout: time.Minute,
		DetectTimeout:  time.Minute,
		ShouldTrip:     testRateTripFuncFalse(1, 1),
	})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		panel.Fail(key)
	}
}

func BenchmarkPanelClosed_Timeout(b *testing.B) {
	panel, err := NewPanel(nil, Options{
		CoolingTimeout: time.Minute,
		DetectTimeout:  time.Minute,
		ShouldTrip:     testRateTripFuncFalse(1, 1),
	})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		panel.Timeout(key)
	}
}

func BenchmarkPanelOpen_IsAllowed(b *testing.B) {
	p, err := NewPanel(nil, Options{
		CoolingTimeout: time.Minute,
		DetectTimeout:  time.Minute,
		ShouldTrip:     testRateTripFuncFalse(1, 1),
	})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	p.(*panel).getBreaker(key).state = Open
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.IsAllowed(key)
	}
}

func BenchmarkPanelOpen_Succeed(b *testing.B) {
	p, err := NewPanel(nil, Options{
		CoolingTimeout: time.Minute,
		DetectTimeout:  time.Minute,
		ShouldTrip:     testRateTripFuncFalse(1, 1),
	})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	p.(*panel).getBreaker(key).state = Open
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Succeed(key)
	}
}

func BenchmarkPanelOpen_Fail(b *testing.B) {
	p, err := NewPanel(nil, Options{
		CoolingTimeout: time.Minute,
		DetectTimeout:  time.Minute,
		ShouldTrip:     testRateTripFuncFalse(1, 1),
	})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	p.(*panel).getBreaker(key).state = Open
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Fail(key)
	}
}

func BenchmarkPanelOpen_Timeout(b *testing.B) {
	p, err := NewPanel(nil, Options{
		CoolingTimeout: time.Minute,
		DetectTimeout:  time.Minute,
		ShouldTrip:     testRateTripFuncFalse(1, 1),
	})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	p.(*panel).getBreaker(key).state = Open
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.Timeout(key)
	}
}

func BenchmarkPanelParallel_Succeed(b *testing.B) {
	panel, err := NewPanel(nil, Options{})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	key2 := "test2"
	key3 := "test3"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			panel.IsAllowed(key)
			panel.IsAllowed(key2)
			panel.IsAllowed(key3)
			panel.Succeed(key)
			panel.Succeed(key2)
			panel.Succeed(key3)
		}
	})
}

func BenchmarkPerPPanelParallel_Succeed(b *testing.B) {
	panel, err := NewPanel(nil, Options{EnableShardP: true})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	key2 := "test2"
	key3 := "test3"
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			panel.IsAllowed(key)
			panel.IsAllowed(key2)
			panel.IsAllowed(key3)
			panel.Succeed(key)
			panel.Succeed(key2)
			panel.Succeed(key3)
		}
	})
}

func BenchmarkPanelOpenParallel_IsAllowed(b *testing.B) {
	p, err := NewPanel(nil, Options{
		CoolingTimeout: time.Minute,
		DetectTimeout:  time.Minute,
		ShouldTrip:     testRateTripFuncFalse(1, 1),
	})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	key2 := "test2"
	key3 := "test3"
	p.(*panel).getBreaker(key).state = Open
	p.(*panel).getBreaker(key2).state = Open
	p.(*panel).getBreaker(key3).state = Open
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.IsAllowed(key)
			p.IsAllowed(key2)
			p.IsAllowed(key3)
		}
	})
}

func BenchmarkPanelParallel2Cores_Succeed(b *testing.B) {
	b.SetParallelism(2)
	p, err := NewPanel(nil, Options{})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	key2 := "test2"
	key3 := "test3"
	p.(*panel).getBreaker(key)
	p.(*panel).getBreaker(key2)
	p.(*panel).getBreaker(key3)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.IsAllowed(key)
			p.IsAllowed(key2)
			p.IsAllowed(key3)
			p.Succeed(key)
			p.Succeed(key2)
			p.Succeed(key3)
		}
	})
}

func BenchmarkPerPPanelParallel2Cores_Succeed(b *testing.B) {
	b.SetParallelism(2)
	p, err := NewPanel(nil, Options{EnableShardP: true})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	key2 := "test2"
	key3 := "test3"
	p.(*panel).getBreaker(key)
	p.(*panel).getBreaker(key2)
	p.(*panel).getBreaker(key3)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.IsAllowed(key)
			p.IsAllowed(key2)
			p.IsAllowed(key3)
			p.Succeed(key)
			p.Succeed(key2)
			p.Succeed(key3)
		}
	})
}

func BenchmarkPanelOpenParallel2Cores_IsAllowed(b *testing.B) {
	b.SetParallelism(2)
	p, err := NewPanel(nil, Options{
		CoolingTimeout: time.Minute,
		DetectTimeout:  time.Minute,
		ShouldTrip:     testRateTripFuncFalse(1, 1),
	})
	if err != nil {
		b.Error(err)
	}
	key := "test"
	key2 := "test2"
	key3 := "test3"
	p.(*panel).getBreaker(key).state = Open
	p.(*panel).getBreaker(key2).state = Open
	p.(*panel).getBreaker(key3).state = Open
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			p.IsAllowed(key)
			p.IsAllowed(key2)
			p.IsAllowed(key3)
		}
	})
}
