package rand

import (
	impl "math/rand"
	"testing"
)

func BenchmarkRandStd(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		impl.Intn(100)
	}
}

func BenchmarkRandParStd(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			impl.Intn(100)
		}
	})
}
