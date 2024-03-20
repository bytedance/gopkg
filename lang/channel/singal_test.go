package channel

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestSignal(t *testing.T) {
	sg := NewSignal()
	var finished int32
	emptyCtx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(emptyCtx)
	for i := 0; i < 10; i++ {
		go func(i int) {
			if i%2 == 0 {
				sg.Wait(emptyCtx)
			} else {
				sg.Wait(cancelCtx)
			}
			atomic.AddInt32(&finished, 1)
		}(i)
	}
	time.Sleep(time.Millisecond * 100)
	cancelFunc()
	for atomic.LoadInt32(&finished) != int32(5) {
		runtime.Gosched()
	}
	sg.Signal()
	for atomic.LoadInt32(&finished) != int32(10) {
		runtime.Gosched()
	}
}

func TestSignalTimeout(t *testing.T) {
	sg := NewSignal(WithSignalTimeout(time.Millisecond * 200))
	go func() {
		time.Sleep(time.Millisecond * 500)
		sg.Signal()
	}()
	begin := time.Now()
	sg.Wait(context.Background())
	cost := time.Since(begin)
	t.Logf("cost=%dms", cost.Milliseconds())
}
