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

	_ "unsafe" // for linkname
)

//go:linkname runtime_fastrand runtime.fastrand
func runtime_fastrand() uint32

//go:nosplit
func fastrandn(n uint32) uint32 {
	// This is similar to Uint32() % n, but faster.
	// See https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32(uint64(runtime_fastrand()) * uint64(n) >> 32)
}

func TestDigest(t *testing.T) {
	d := NewDefault()
	for size := 0; size <= 1024; size++ {
		data := make([]byte, size)
		for i := range data {
			data[i] = byte(fastrandn(256))
		}
		// Random write small data.
		var r int
		if size == 0 {
			r = 0
		} else {
			r = int(fastrandn(uint32(len(data))))
		}
		d.Write(data[:r])
		d.Write(data[r:])
		if d.Sum64() != Sum64(data) {
			t.Fatal(size, d.Sum64(), Sum64(data))
		}
		d.Reset()
	}

	largedata := make([]byte, 1024*1024)
	for i := range largedata {
		largedata[i] = byte(fastrandn(256))
	}

	var a, b int
	digest := NewDefault()
	partsizelimit := 300
	for {
		if len(largedata)-a < 300 {
			b = len(largedata) - a
		} else {
			b = int(fastrandn(uint32(partsizelimit)))
		}
		digest.Write(largedata[a : a+b])
		if Sum64(largedata[:a+b]) != digest.Sum64() {
			t.Fatal(a, b)
		}
		a += b
		if a == len(largedata) {
			break
		}
	}
}

func BenchmarkDigest(b *testing.B) {
	sizes := []int{33, 64, 96, 128, 129, 240, 241,
		512, 1024, 10 * 1024,
	}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d", size), func(b *testing.B) {
			b.SetBytes(int64(size))
			var acc uint64
			data := make([]byte, size)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				d := NewDefault()
				d.Write(data)
				acc = d.Sum64()
			}
			runtime.KeepAlive(acc)
		})
	}
}
