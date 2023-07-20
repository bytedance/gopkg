package ptuner

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestTuner(t *testing.T) {
	old := runtime.GOMAXPROCS(2)
	defer runtime.GOMAXPROCS(old)

	err := Tuning(
		WithMaxProcs(5),
		//WithTuningLimit(3),
		WithTuningFrequency(time.Second),
	)
	if err != nil {
		t.FailNow()
	}

	var stop int32
	for i := 0; i < 10; i++ {
		sum := 0
		go func(id int) {
			for atomic.LoadInt32(&stop) == 0 {
				sum += userFunc(id*10 + 1)
			}
		}(i)
	}

	time.Sleep(time.Second * 20)
	atomic.StoreInt32(&stop, 1)
	time.Sleep(time.Second * 20)
}
