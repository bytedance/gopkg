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
	defaultMinSize        = 1
)

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

// WithSize define the size of channel. If channel is full, it will block.
// It conflicts with WithNonBlock option.
func WithSize(size int) Option {
	return func(c *channel) {
		// with non block mode, no need to change size
		if size >= defaultMinSize && !c.nonblock {
			c.size = size
		}
	}
}

// WithNonBlock will set channel to non-blocking Mode.
// The input channel will not block for any cases.
func WithNonBlock() Option {
	return func(c *channel) {
		c.nonblock = true
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

var (
	_ Channel = (*channel)(nil)
)

// Channel is a safe and feature-rich alternative for Go chan struct
type Channel interface {
	// Input send value to Output channel. If channel is closed, do nothing and will not panic.
	Input(v interface{})
	// Output return a read-only native chan for consumer.
	Output() <-chan interface{}
	// Len return the count of un-consumed items.
	Len() int
	// Stats return the produced and consumed count.
	Stats() (produced uint64, consumed uint64)
	// Close closed the output chan. If channel is not closed explicitly, it will be closed when it's finalized.
	Close()
}

// channelWrapper use to detect user never hold the reference of Channel object, and runtime will help to close channel implicitly.
type channelWrapper struct {
	Channel
}

// channel implements a safe and feature-rich channel struct for the real world.
type channel struct {
	size             int
	state            int32
	consumer         chan interface{}
	nonblock         bool // non blocking mode
	timeout          time.Duration
	timeoutCallback  func(interface{})
	producerThrottle Throttle
	consumerThrottle Throttle
	throttleWindow   time.Duration
	// statistics
	produced uint64 // item already been insert into buffer
	consumed uint64 // item already been sent into Output chan
	// buffer
	buffer     *list.List // TODO: use high perf queue to reduce GC here
	bufferCond *sync.Cond
	bufferLock sync.Mutex
}

// New create a new channel.
func New(opts ...Option) Channel {
	c := new(channel)
	c.size = defaultMinSize
	c.throttleWindow = defaultThrottleWindow
	c.bufferCond = sync.NewCond(&c.bufferLock)
	for _, opt := range opts {
		opt(c)
	}
	c.consumer = make(chan interface{})
	c.buffer = list.New()
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
	c.bufferLock.Lock()
	defer c.bufferLock.Unlock()
	// Close function only notify Input/consume goroutine to close gracefully
	c.bufferCond.Broadcast()
}

func (c *channel) isClosed() bool {
	return atomic.LoadInt32(&c.state) < 0
}

func (c *channel) Input(v interface{}) {
	if c.isClosed() {
		return
	}

	// prepare item
	it := item{value: v}
	if c.timeout > 0 {
		it.deadline = time.Now().Add(c.timeout)
	}

	// only check throttle function in blocking mode
	if !c.nonblock {
		if c.throttling(c.producerThrottle) {
			// closed
			return
		}
	}

	// enqueue buffer
	c.bufferLock.Lock()
	if !c.nonblock {
		// only check length with blocking mode
		for c.buffer.Len() >= c.size {
			// wait for consuming
			c.bufferCond.Wait()
			if c.isClosed() {
				// blocking send a closed channel should return directly
				c.bufferLock.Unlock()
				return
			}
		}
	}
	c.enqueueBuffer(it)
	atomic.AddUint64(&c.produced, 1)
	c.bufferCond.Signal() // use Signal because only 1 goroutine wait for cond
	c.bufferLock.Unlock()
}

func (c *channel) Output() <-chan interface{} {
	return c.consumer
}

func (c *channel) Len() int {
	produced, consumed := c.Stats()
	l := produced - consumed
	return int(l)
}

func (c *channel) Stats() (uint64, uint64) {
	produced, consumed := atomic.LoadUint64(&c.produced), atomic.LoadUint64(&c.consumed)
	return produced, consumed
}

// consume used to process input buffer
func (c *channel) consume() {
	for {
		// check throttle
		if c.throttling(c.consumerThrottle) {
			// closed
			return
		}

		// dequeue buffer
		c.bufferLock.Lock()
		for c.buffer.Len() == 0 {
			if c.isClosed() {
				close(c.consumer)               // close consumer
				atomic.StoreInt32(&c.state, -2) // -2 means closed totally
				c.bufferLock.Unlock()
				return
			}
			c.bufferCond.Wait()
		}
		it, ok := c.dequeueBuffer()
		c.bufferCond.Broadcast() // use Broadcast because there will be more than 1 goroutines wait for cond
		c.bufferLock.Unlock()
		if !ok {
			// in fact, this case will never happen
			continue
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

func (c *channel) throttling(throttle Throttle) (closed bool) {
	if throttle == nil {
		return
	}
	throttled := throttle(c)
	if !throttled {
		return
	}
	ticker := time.NewTicker(c.throttleWindow)
	defer ticker.Stop()

	closed = c.isClosed()
	for throttled && !closed {
		<-ticker.C
		throttled, closed = throttle(c), c.isClosed()
	}
	return closed
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
