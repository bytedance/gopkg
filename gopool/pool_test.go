package gopool

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

const benchmarkTimes = 10000

func testFunc() {
	for i := 0; i < benchmarkTimes; i++ {
		rand.Intn(benchmarkTimes)
	}
}

func testPanicFunc() {
	panic("test")
}

func TestPool(t *testing.T) {
	p := NewPool("test", 100, NewConfig())
	var n int32
	var wg sync.WaitGroup
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		p.Go(func() {
			defer wg.Done()
			atomic.AddInt32(&n, 1)
		})
	}
	wg.Wait()
	if n != 2000 {
		t.Error(n)
	}
}

func TestPoolPanic(t *testing.T) {
	p := NewPool("test", 100, NewConfig())
	p.Go(testPanicFunc)
}

func BenchmarkPool(b *testing.B) {
	fmt.Println(runtime.GOMAXPROCS(0))
	config := NewConfig()
	config.ScaleThreshold = 1
	p := NewPool("benchmark", int32(runtime.GOMAXPROCS(0)), config)
	var wg sync.WaitGroup
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(benchmarkTimes)
		for j := 0; j < benchmarkTimes; j++ {
			p.Go(func() {
				testFunc()
				wg.Done()
			})
		}
		wg.Wait()
	}
	p.Close()
}

func BenchmarkGo(b *testing.B) {
	var wg sync.WaitGroup
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(benchmarkTimes)
		for j := 0; j < benchmarkTimes; j++ {
			go func() {
				testFunc()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}
