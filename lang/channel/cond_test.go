package channel

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestCond(t *testing.T) {
	cd := NewCond()
	var finished int32
	emptyCtx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(emptyCtx)
	for i := 0; i < 10; i++ {
		go func(i int) {
			if i%2 == 0 {
				cd.Wait(emptyCtx)
			} else {
				cd.Wait(cancelCtx)
			}
			atomic.AddInt32(&finished, 1)
		}(i)
	}
	time.Sleep(time.Millisecond * 100)
	cancelFunc()
	for atomic.LoadInt32(&finished) != int32(5) {
		runtime.Gosched()
	}
	cd.Signal()
	for atomic.LoadInt32(&finished) != int32(6) {
		runtime.Gosched()
	}
	cd.Signal()
	for atomic.LoadInt32(&finished) != int32(7) {
		runtime.Gosched()
	}
	cd.Broadcast()
	cd.Signal()
	for atomic.LoadInt32(&finished) != int32(10) {
		runtime.Gosched()
	}
}

func TestCondTimeout(t *testing.T) {
	cd := NewCond(WithCondTimeout(time.Millisecond * 200))
	go func() {
		time.Sleep(time.Millisecond * 500)
		cd.Broadcast()
	}()
	begin := time.Now()
	cd.Wait(context.Background())
	cost := time.Since(begin)
	t.Logf("cost=%dms", cost.Milliseconds())
}

func BenchmarkChanCond(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ch := make(chan struct{})
		go func() {
			time.Sleep(time.Millisecond)
			close(ch)
		}()
		select {
		case <-ch:
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func BenchmarkCond(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cd := NewCond(WithCondTimeout(10 * time.Millisecond))
		go func() {
			time.Sleep(time.Millisecond)
			cd.Signal()
		}()
		cd.Wait(context.Background())
	}
}
