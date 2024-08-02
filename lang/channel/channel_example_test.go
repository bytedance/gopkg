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
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type request struct {
	Id      int
	Latency time.Duration
	Done    chan struct{}
}

type response struct {
	Id int
}

var taskPool Channel

func Service1(req *request) {
	taskPool.Input(req)
	return
}

func Service2(req *request) (*response, error) {
	if req.Latency > 0 {
		time.Sleep(req.Latency)
	}
	return &response{Id: req.Id}, nil
}

func TestNetworkIsolationOrDownstreamBlock(t *testing.T) {
	taskPool = New(
		WithNonBlock(),
		WithTimeout(time.Millisecond*10),
	)
	defer taskPool.Close()
	var responded int32
	go func() {
		// task worker
		for task := range taskPool.Output() {
			req := task.(*request)
			done := make(chan struct{})
			go func() {
				_, _ = Service2(req)
				close(done)
			}()
			select {
			case <-time.After(time.Millisecond * 100):
			case <-done:
				atomic.AddInt32(&responded, 1)
			}
		}
	}()

	start := time.Now()
	for i := 1; i <= 100; i++ {
		req := &request{Id: i}
		if i > 50 && i <= 60 { // suddenly have network issue for 10 requests
			req.Latency = time.Hour
		}
		Service1(req)
	}
	cost := time.Now().Sub(start)
	assert.True(t, cost < time.Millisecond*10)               // Service1 should not block
	time.Sleep(time.Millisecond * 1500)                      // wait all tasks finished
	assert.Equal(t, int32(50), atomic.LoadInt32(&responded)) // 50 success and 10 timeout and 40 discard
}

func TestCPUHeavy(t *testing.T) {
	runtime.GOMAXPROCS(1)
	var concurrency int32
	taskPool = New(
		WithNonBlock(),
		WithThrottle(nil, func(c Channel) bool {
			return atomic.LoadInt32(&concurrency) > 10
		}),
	)
	defer taskPool.Close()
	var responded int32
	go func() {
		// task worker
		for task := range taskPool.Output() {
			req := task.(*request)
			t.Logf("NumGoroutine: %d", runtime.NumGoroutine())
			go func() {
				curConcurrency := atomic.AddInt32(&concurrency, 1)
				defer atomic.AddInt32(&concurrency, -1)
				if curConcurrency > 10 {
					// concurrency too high, reuqest faild
					return
				}

				atomic.AddInt32(&responded, 1)
				if req.Id >= 11 && req.Id <= 20 {
					start := time.Now()
					for x := uint64(0); ; x++ {
						if x%1000 == 0 {
							if time.Now().Sub(start) >= 100*time.Millisecond {
								return
							}
						}
					}
				}
			}()
		}
	}()

	start := time.Now()
	for i := 1; i <= 100; i++ {
		req := &request{Id: i}
		Service1(req)
	}
	cost := time.Now().Sub(start)
	assert.True(t, cost < time.Millisecond*10)            // Service1 should not block
	time.Sleep(time.Second * 2)                           // wait all tasks finished
	t.Logf("responded: %d", atomic.LoadInt32(&responded)) // most tasks success
}
