package ptuner

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// userFunc include cpu and gc work
//
//go:noinline
func userFunc(n int) (ret int) {
	if n == 0 {
		return 0
	}
	sum := make([]int, n)
	for i := 0; i < n; i++ {
		sum[i] = userFunc(i / 2)
	}
	for _, x := range sum {
		ret += x
	}
	return ret
}

func TestMetrics(t *testing.T) {
	old := runtime.GOMAXPROCS(4)
	defer runtime.GOMAXPROCS(old)

	var stop int32
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sum := 0
			for atomic.LoadInt32(&stop) == 0 {
				sum += userFunc(i * 10)
			}
			t.Logf("goroutine[%d] exit", i)
		}(i)
	}

	ra := newRuntimeAnalyzer()
	for i := 0; ; i++ {
		time.Sleep(time.Second)
		schedLatency, cpuPercent := ra.Analyze()
		t.Logf("schedLatency=%.2fms cpuPercent=%.2f%%", schedLatency*1000, cpuPercent*100)

		if i == 5 {
			atomic.StoreInt32(&stop, 1)
			t.Logf("stop background goroutines")
		}
	}
}
