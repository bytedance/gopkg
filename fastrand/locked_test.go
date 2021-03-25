package rand

import (
	"sync"
	"testing"
)

func BenchmarkLocked(b *testing.B) {
	r := NewLocked()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Intn(100)
	}
}

func BenchmarkLockedPar(b *testing.B) {
	r := NewLocked()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			r.Intn(100)
		}
	})
}

func TestIntn(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = Intn(101)
		}()
	}

	wg.Wait()
}
