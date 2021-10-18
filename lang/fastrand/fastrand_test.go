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

	p := make([]byte, 1000)
	n, err := Read(p)
	if n != len(p) || err != nil || (p[0] == 0 && p[1] == 0 && p[2] == 0) {
		t.Fatal()
	}

	a := Perm(100)
	for i := range a {
		var find bool
		for _, v := range a {
			if v == i {
				find = true
			}
		}
		if !find {
			t.Fatal()
		}
	}

	Shuffle(len(a), func(i, j int) {
		a[i], a[j] = a[j], a[i]
	})
	for i := range a {
		var find bool
		for _, v := range a {
			if v == i {
				find = true
			}
		}
		if !find {
			t.Fatal()
		}
	}
}

func BenchmarkSingleCore(b *testing.B) {
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

	b.Run("math/rand-Read1000", func(b *testing.B) {
		p := make([]byte, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			rand.Read(p)
		}
	})
	b.Run("fast-rand-Read1000", func(b *testing.B) {
		p := make([]byte, 1000)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Read(p)
		}
	})

}

func BenchmarkMultipleCore(b *testing.B) {
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

	b.Run("math/rand-Read1000", func(b *testing.B) {
		p := make([]byte, 1000)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				rand.Read(p)
			}
		})
	})
	b.Run("fast-rand-Read1000", func(b *testing.B) {
		p := make([]byte, 1000)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				Read(p)
			}
		})
	})
}
