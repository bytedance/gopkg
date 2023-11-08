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
	"container/list"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultThrottleWindow = time.Millisecond * 100
	defaultSize           = 0
)

// terminalSig is special item to define system message
var terminalSig interface{} = struct{}{}

type item struct {
	value    interface{}
	deadline time.Time
}

// IsExpired check is item exceed deadline, zero means non-expired
func (i item) IsExpired() bool {
	if i.deadline.IsZero() {
		return false
	}
	return time.Now().After(i.deadline)
}

// Option define channel Option
type Option func(c *channel)

// Throttle define channel Throttle function
type Throttle func(c Channel) bool

// WithSize define the size of channel.
// It conflicts with WithNonBlock option.
func WithSize(size int) Option {
	return func(c *channel) {
		// with non block mode, no need to change size
		if !c.nonblock {
			c.size = size
		}
	}
}

// WithNonBlock will set channel to non-blocking Mode.
// The input channel will not block for any cases.
func WithNonBlock() Option {
	return func(c *channel) {
		c.nonblock = true
		c.size = 1024
	}
}

// WithTimeout sets the expiration time of each channel item.
// If the item not consumed in timeout duration, it will be aborted.
func WithTimeout(timeout time.Duration) Option {
	return func(c *channel) {
		c.timeout = timeout
	}
}

// WithTimeoutCallback sets callback function when item hit timeout.
func WithTimeoutCallback(timeoutCallback func(interface{})) Option {
	return func(c *channel) {
		c.timeoutCallback = timeoutCallback
	}
}

// WithThrottle sets both producerThrottle and consumerThrottle
// If producerThrottle throttled, it input channel will be blocked(if using blocking mode).
// If consumerThrottle throttled, it output channel will be blocked.
func WithThrottle(producerThrottle, consumerThrottle Throttle) Option {
	return func(c *channel) {
		if c.producerThrottle == nil {
			c.producerThrottle = producerThrottle
		} else {
			prevChecker := c.producerThrottle
			c.producerThrottle = func(c Channel) bool {
				return prevChecker(c) && producerThrottle(c)
			}
		}
		if c.consumerThrottle == nil {
			c.consumerThrottle = consumerThrottle
		} else {
			prevChecker := c.consumerThrottle
			c.consumerThrottle = func(c Channel) bool {
				return prevChecker(c) && consumerThrottle(c)
			}
		}
	}
}

// WithThrottleWindow sets the interval time for throttle function checking.
func WithThrottleWindow(window time.Duration) Option {
	return func(c *channel) {
		c.throttleWindow = window
	}
}

// WithRateThrottle is a helper function to control producer and consumer process rate.
// produceRate and consumeRate mean how many item could be processed in one second, aka TPS.
func WithRateThrottle(produceRate, consumeRate int) Option {
	// throttle function will be called sequentially
	producedMax := uint64(produceRate)
	consumedMax := uint64(consumeRate)
	var producedBegin, consumedBegin uint64
	var producedTS, consumedTS int64
	return WithThrottle(func(c Channel) bool {
		ts := time.Now().Unix() // in second
		produced, _ := c.Stats()
		if producedTS != ts {
			// move to a new second, so store the current process as beginning value
			producedBegin = produced
			producedTS = ts
			return false
		}
		// get the value of beginning
		producedDiff := produced - producedBegin
		return producedMax > 0 && producedMax < producedDiff
	}, func(c Channel) bool {
		ts := time.Now().Unix() // in second
		_, consumed := c.Stats()
		if consumedTS != ts {
			// move to a new second, so store the current process as beginning value
			consumedBegin = consumed
			consumedTS = ts
			return false
		}
		// get the value of beginning
		consumedDiff := consumed - consumedBegin
		return consumedMax > 0 && consumedMax < consumedDiff
	})
}

var _ Channel = (*channel)(nil)

type Channel interface {
	// Input return a native chan for produce task
	Input() chan interface{}
	// Output return a native chan for consume task
	Output() chan interface{}
	// Len return the count of un-consumed tasks
	Len() int
	// Stats return the produced and consumed count
	Stats() (produced uint64, consumed uint64)
	// Close will close the producer and consumer goroutines gracefully
	Close()
}

// channelWrapper use to detect user never hold the reference of channel object, and we need to close channel implicitly.
type channelWrapper struct {
	Channel
}

// channel implements a safe and feature-rich channel struct for the real world.
type channel struct {
	size             int
	state            int32
	producer         chan interface{}
	consumer         chan interface{}
	timeout          time.Duration
	timeoutCallback  func(interface{})
	producerThrottle Throttle
	consumerThrottle Throttle
	throttleWindow   time.Duration
	// statistics
	produced uint64
	consumed uint64
	// non blocking mode
	nonblock bool
	// buffer
	buffer     *list.List // TODO: use high perf queue to reduce GC here
	bufferCond *sync.Cond
	bufferLock sync.Mutex
}

// New create a new channel.
func New(opts ...Option) Channel {
	c := new(channel)
	c.size = defaultSize
	c.throttleWindow = defaultThrottleWindow
	c.bufferCond = sync.NewCond(&c.bufferLock)
	for _, opt := range opts {
		opt(c)
	}
	c.producer = make(chan interface{}, c.size)
	c.consumer = make(chan interface{})
	c.buffer = list.New()
	go c.produce()
	go c.consume()

	// register finalizer for wrapper of channel
	cw := &channelWrapper{c}
	runtime.SetFinalizer(cw, func(obj *channelWrapper) {
		// it's ok to call Close again if user already closed the channel
		obj.Close()
	})
	return cw
}

// Close will close the producer and consumer goroutines gracefully
func (c *channel) Close() {
	if !atomic.CompareAndSwapInt32(&c.state, 0, -1) {
		return
	}
	// empty buffer
	c.bufferLock.Lock()
	c.buffer.Init() // clear
	c.bufferLock.Unlock()
	c.bufferCond.Broadcast()
	c.producer <- terminalSig
}

func (c *channel) isClosed() bool {
	return atomic.LoadInt32(&c.state) < 0
}

// Input return a native chan for produce task
func (c *channel) Input() chan interface{} {
	return c.producer
}

// Output return a native chan for consume task
func (c *channel) Output() chan interface{} {
	return c.consumer
}

// Len return the count of un-consumed tasks.
func (c *channel) Len() int {
	produced, consumed := c.Stats()
	l := produced - consumed
	return int(l)
}

func (c *channel) Stats() (uint64, uint64) {
	produced, consumed := atomic.LoadUint64(&c.produced), atomic.LoadUint64(&c.consumed)
	return produced, consumed
}

// produce used to process input channel
func (c *channel) produce() {
	capacity := c.size
	if c.size == 0 {
		capacity = 1
	}
	for p := range c.producer {
		// only check throttle function in blocking mode
		if !c.nonblock {
			c.throttling(c.producerThrottle)
		}

		// produced
		atomic.AddUint64(&c.produced, 1)
		// prepare item
		it := item{value: p}
		if c.timeout > 0 {
			it.deadline = time.Now().Add(c.timeout)
		}
		// enqueue buffer
		c.bufferLock.Lock()
		c.enqueueBuffer(it)
		c.bufferCond.Signal()
		if !c.nonblock {
			for c.buffer.Len() >= capacity {
				c.bufferCond.Wait()
			}
		}
		c.bufferLock.Unlock()

		if p == terminalSig { // graceful shutdown
			close(c.producer)
			return
		}
	}
}

// consume used to process output channel
func (c *channel) consume() {
	for {
		// check throttle
		c.throttling(c.consumerThrottle)

		// dequeue buffer
		c.bufferLock.Lock()
		for c.buffer.Len() == 0 {
			c.bufferCond.Wait()
		}
		it, ok := c.dequeueBuffer()
		c.bufferLock.Unlock()
		c.bufferCond.Signal()
		if !ok {
			// in fact, this case will never happen
			continue
		}

		// graceful shutdown
		if it.value == terminalSig {
			atomic.AddUint64(&c.consumed, 1)
			close(c.consumer)
			atomic.StoreInt32(&c.state, -2)
			return
		}

		// check expired
		if it.IsExpired() {
			if c.timeoutCallback != nil {
				c.timeoutCallback(it.value)
			}
			atomic.AddUint64(&c.consumed, 1)
			continue
		}
		// consuming, if block here means consumer is busy
		c.consumer <- it.value
		atomic.AddUint64(&c.consumed, 1)
	}
}

func (c *channel) throttling(throttle Throttle) {
	if throttle == nil {
		return
	}
	throttled := throttle(c)
	if !throttled {
		return
	}
	ticker := time.NewTicker(c.throttleWindow)
	defer ticker.Stop()

	for throttled && !c.isClosed() {
		<-ticker.C
		throttled = throttle(c)
	}
}

func (c *channel) enqueueBuffer(it item) {
	c.buffer.PushBack(it)
}

func (c *channel) dequeueBuffer() (it item, ok bool) {
	bi := c.buffer.Front()
	if bi == nil {
		return it, false
	}
	c.buffer.Remove(bi)

	it = bi.Value.(item)
	return it, true
}
