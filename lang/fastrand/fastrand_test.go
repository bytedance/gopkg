// Copyright 2021 ByteDance Inc.
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

package fastrand

import (
	"math/rand"
	"testing"
)

func TestAll(t *testing.T) {
	_ = Uint32()

	bytes := make([]byte, 1000)
	_, _ = Read(bytes)
}

func BenchmarkSingleCore(b *testing.B) {
	b.Run("math/rand-Int31n(5)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rand.Int31n(5)
		}
	})
	b.Run("fast-rand-Int31n(5)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Int31n(5)
		}
	})

	b.Run("math/rand-Int63n(5)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rand.Int63n(5)
		}
	})
	b.Run("fast-rand-Int63n(5)", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Int63n(5)
		}
	})

	b.Run("math/rand-Uint32()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rand.Uint32()
		}
	})
	b.Run("fast-rand-Uint32()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Uint32()
		}
	})
	b.Run("math/rand-Uint64()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = rand.Uint64()
		}
	})

	b.Run("fast-rand-Uint64()", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Uint64()
		}
	})

}

func BenchmarkMultipleCore(b *testing.B) {
	b.Run("math/rand-Int31n(5)", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = rand.Int31n(5)
			}
		})
	})
	b.Run("fast-rand-Int31n(5)", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = Int31n(5)
			}
		})
	})

	b.Run("math/rand-Int63n(5)", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = rand.Int63n(5)
			}
		})
	})
	b.Run("fast-rand-Int63n(5)", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = Int63n(5)
			}
		})
	})

	b.Run("math/rand-Float32()", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = rand.Float32()
			}
		})
	})
	b.Run("fast-rand-Float32()", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = Float32()
			}
		})
	})

	b.Run("math/rand-Uint32()", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = rand.Uint32()
			}
		})
	})
	b.Run("fast-rand-Uint32()", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = Uint32()
			}
		})
	})

	b.Run("math/rand-Uint64()", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = rand.Uint64()
			}
		})
	})
	b.Run("fast-rand-Uint64()", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = Uint64()
			}
		})
	})
}
