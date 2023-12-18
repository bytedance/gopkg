// Copyright 2023 ByteDance Inc.
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

package channel

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func tlogf(t *testing.T, format string, args ...interface{}) {
	t.Log(fmt.Sprintf("[%v] %s", time.Now().UTC(), fmt.Sprintf(format, args...)))
}

//go:noinline
func factorial(x int) int {
	if x <= 1 {
		return x
	}
	return x * factorial(x-1)
}

var benchSizes = []int{1, 10, 100, 1000, -1}

func BenchmarkNativeChan(b *testing.B) {
	for _, size := range benchSizes {
		if size < 0 {
			continue
		}
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			ch := make(chan interface{}, size)
			b.RunParallel(func(pb *testing.PB) {
				n := 0
				for pb.Next() {
					n++
					ch <- n
					<-ch
				}
			})
		})
	}
}

func BenchmarkChannel(b *testing.B) {
	for _, size := range benchSizes {
		b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
			var ch Channel
			if size < 0 {
				ch = New(WithNonBlock())
			} else {
				ch = New(WithSize(size))
			}
			defer ch.Close()
			b.RunParallel(func(pb *testing.PB) {
				n := 0
				for pb.Next() {
					n++
					ch.Input() <- n
					<-ch.Output()
				}
			})
		})
	}
}

func TestChannelDefaultSize(t *testing.T) {
	ch := New()
	defer ch.Close()

	for i := 1; i <= 10; i++ {
		ch.Input() <- i
		t.Logf("put %d", i)
		x := <-ch.Output()
		t.Logf("get %d", x)
		assert.Equal(t, i, x)
	}
	ch.Input() <- 0 // wait for be consumed
	t.Logf("put 0-1")
	ch.Input() <- 0 // wait for be buffered
	t.Logf("put 0-2")
	timeout := false
	select {
	case ch.Input() <- 0: // block
	case <-time.After(time.Millisecond * 10):
		timeout = true
	}
	assert.True(t, timeout)
}

func TestChannelClose(t *testing.T) {
	beginGs := runtime.NumGoroutine()
	ch := New()
	afterGs := runtime.NumGoroutine()
	assert.Equal(t, 2, afterGs-beginGs)
	var exit int32
	go func() {
		for _ = range ch.Output() {
		}
		atomic.AddInt32(&exit, 1)
	}()
	for i := 1; i <= 20; i++ {
		ch.Input() <- i
	}
	ch.Close()
	for runtime.NumGoroutine() > beginGs {
		tlogf(t, "num goroutines: %d, beginGs: %d", runtime.NumGoroutine(), beginGs)
		runtime.Gosched()
	}
	<-ch.Output() // never block
	assert.Equal(t, int32(1), atomic.LoadInt32(&exit))
}

func TestChannelTimeout(t *testing.T) {
	ch := New(
		WithTimeout(time.Millisecond*50),
		WithSize(1024),
	)
	defer ch.Close()

	go func() {
		for i := 1; i <= 20; i++ {
			ch.Input() <- i
		}
	}()
	var total int32
	go func() {
		for c := range ch.Output() {
			id := c.(int)
			if id >= 10 {
				time.Sleep(time.Millisecond * 100)
			}
			atomic.AddInt32(&total, 1)
		}
	}()
	time.Sleep(time.Second)
	// success task: id in [1, 11]
	// note that task with id=11 also will be consumed since it already checked.
	assert.Equal(t, int32(11), atomic.LoadInt32(&total))
}

func TestChannelConsumerInflightLimit(t *testing.T) {
	var inflight int32
	var limit int32 = 10
	var total = 20
	ch := New(
		WithThrottle(nil, func(c Channel) bool {
			return atomic.LoadInt32(&inflight) >= limit
		}),
	)
	defer ch.Close()

	var wg sync.WaitGroup
	go func() {
		for c := range ch.Output() {
			atomic.AddInt32(&inflight, 1)
			id := c.(int)
			tlogf(t, "consumer=%d started", id)
			go func() {
				defer atomic.AddInt32(&inflight, -1)
				defer wg.Done()
				time.Sleep(time.Second)
				tlogf(t, "consumer=%d finished", id)
			}()
		}
	}()

	now := time.Now()
	for i := 1; i <= total; i++ {
		wg.Add(1)
		id := i
		ch.Input() <- id
		tlogf(t, "producer=%d finished", id)
		time.Sleep(time.Millisecond * 10)
	}
	wg.Wait()
	duration := time.Now().Sub(now)
	assert.Equal(t, 2, int(duration.Seconds()))
}

func TestChannelProducerSpeedLimit(t *testing.T) {
	var total = 15
	ch := New(WithSize(0))
	defer ch.Close()

	go func() {
		for c := range ch.Output() {
			id := c.(int)
			time.Sleep(time.Millisecond * 100)
			tlogf(t, "consumer=%d finished", id)
		}
	}()

	now := time.Now()
	for i := 1; i <= total; i++ {
		id := i
		ch.Input() <- id
		tlogf(t, "producer=%d finished", id)
	}
	duration := time.Now().Sub(now)
	assert.Equal(t, 1, int(duration.Seconds()))
}

func TestChannelProducerNoLimit(t *testing.T) {
	var total = 100
	ch := New(WithSize(1000))
	defer ch.Close()

	go func() {
		for c := range ch.Output() {
			id := c.(int)
			time.Sleep(time.Millisecond * 100)
			tlogf(t, "consumer=%d finished", id)
		}
	}()

	now := time.Now()
	for i := 1; i <= total; i++ {
		id := i
		ch.Input() <- id
	}
	duration := time.Now().Sub(now)
	assert.Equal(t, 0, int(duration.Seconds()))
}

func TestChannelGoroutinesThrottle(t *testing.T) {
	goroutineChecker := func(maxGoroutines int) Throttle {
		return func(c Channel) bool {
			tlogf(t, "%d goroutines", runtime.NumGoroutine())
			return runtime.NumGoroutine() > maxGoroutines
		}
	}
	var total = 1000
	throttle := goroutineChecker(100)
	ch := New(WithThrottle(throttle, throttle), WithThrottleWindow(time.Millisecond*100))
	var wg sync.WaitGroup
	go func() {
		for c := range ch.Output() {
			id := c.(int)
			go func() {
				time.Sleep(time.Millisecond * 100)
				tlogf(t, "consumer=%d finished", id)
				wg.Done()
			}()
		}
	}()

	for i := 1; i <= total; i++ {
		wg.Add(1)
		id := i
		ch.Input() <- id
		tlogf(t, "producer=%d finished", id)
		runtime.Gosched()
	}
	wg.Wait()
}

func TestChannelNoConsumer(t *testing.T) {
	ch := New()

	var sum int32
	go func() {
		for i := 1; i <= 20; i++ {
			ch.Input() <- i
			tlogf(t, "producer=%d finished", i)
			atomic.AddInt32(&sum, 1)
		}
	}()
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, int32(2), atomic.LoadInt32(&sum))
}

func TestChannelOneSlowTask(t *testing.T) {
	ch := New(WithTimeout(time.Millisecond*500), WithSize(0))
	defer ch.Close()

	var total int32
	go func() {
		for c := range ch.Output() {
			id := c.(int)
			if id == 10 {
				time.Sleep(time.Second)
			}
			atomic.AddInt32(&total, 1)
		}
	}()

	for i := 1; i <= 20; i++ {
		ch.Input() <- i
		tlogf(t, "producer=%d finished", i)
	}
	time.Sleep(time.Second)
	assert.Equal(t, int32(19), atomic.LoadInt32(&total))
}

func TestChannelProduceRateControl(t *testing.T) {
	produceMaxRate := 100
	ch := New(
		WithRateThrottle(produceMaxRate, 0),
	)
	defer ch.Close()

	go func() {
		for c := range ch.Output() {
			id := c.(int)
			tlogf(t, "consumed: %d", id)
		}
	}()
	begin := time.Now()
	for i := 1; i <= 500; i++ {
		ch.Input() <- i
	}
	cost := time.Now().Sub(begin)
	tlogf(t, "Cost %dms", cost.Milliseconds())
}

func TestChannelConsumeRateControl(t *testing.T) {
	ch := New(
		WithRateThrottle(0, 100),
	)
	defer ch.Close()

	go func() {
		for c := range ch.Output() {
			id := c.(int)
			tlogf(t, "consumed: %d", id)
		}
	}()
	begin := time.Now()
	for i := 1; i <= 500; i++ {
		ch.Input() <- i
	}
	cost := time.Now().Sub(begin)
	tlogf(t, "Cost %dms", cost.Milliseconds())
}

func TestChannelNonBlock(t *testing.T) {
	ch := New(WithNonBlock())
	defer ch.Close()

	begin := time.Now()
	for i := 1; i <= 10000; i++ {
		ch.Input() <- i
		tlogf(t, "producer=%d finished", i)
	}
	cost := time.Now().Sub(begin)
	tlogf(t, "Cost %dms", cost.Milliseconds())
}

func TestAvoidGoroutineLeak(t *testing.T) {
	// Default channel is safe
	recvCh := New()
	var wg sync.WaitGroup
	wg.Add(1)
	// producer
	go func() {
		time.Sleep(time.Millisecond * 100) // RPC Call
		recvCh.Input() <- 1
		wg.Done()
	}()
	// consumer
	select {
	case <-recvCh.Output():
	case <-time.After(time.Millisecond * 50):
	}
	wg.Wait() // goroutine exit
}

func TestFastRecoverConsumer(t *testing.T) {
	var consumed int32
	var aborted int32
	timeout := time.Second * 1
	ch := New(
		WithNonBlock(),
		WithTimeout(timeout),
		WithTimeoutCallback(func(i interface{}) {
			atomic.AddInt32(&aborted, 1)
		}),
	)
	defer ch.Close()

	// consumer
	go func() {
		for c := range ch.Output() {
			id := c.(int)
			t.Logf("consumed: %d", id)
			time.Sleep(time.Millisecond * 100)
			atomic.AddInt32(&consumed, 1)
		}
	}()

	// producer
	// faster than consumer's ability
	for i := 1; i <= 20; i++ {
		ch.Input() <- i
		time.Sleep(time.Millisecond * 10)
	}
	for (atomic.LoadInt32(&consumed) + atomic.LoadInt32(&aborted)) != 20 {
		runtime.Gosched()
	}
	assert.True(t, aborted > 5)
	consumed = 0
	aborted = 0
	// quick recover consumer
	for i := 1; i <= 10; i++ {
		ch.Input() <- i
		time.Sleep(time.Millisecond * 10)
	}
	for atomic.LoadInt32(&consumed) != 10 {
		runtime.Gosched()
	}
	// all consumed
}
