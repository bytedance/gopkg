package runtimex

import "testing"

func BenchmarkFastrand(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Fastrand()
	}
}
