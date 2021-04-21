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

package wyhash

import (
	"fmt"
	"runtime"
	"testing"
)

func BenchmarkWyhash(b *testing.B) {
	sizes := []int{17, 21, 24, 29, 32,
		33, 64, 69, 96, 97, 128, 129, 240, 241,
		512, 1024, 100 * 1024,
	}

	for size := 0; size <= 16; size++ {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			var (
				x    uint64
				data = string(make([]byte, size))
			)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				x = Sum64String(data)
			}
			runtime.KeepAlive(x)
		})
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			var x uint64
			data := string(make([]byte, size))
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				x = Sum64String(data)
			}
			runtime.KeepAlive(x)
		})
	}
}
