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

package xxhash3

import (
	"runtime"
	"testing"
)

type benchTask struct {
	name      string
	version   int
	action    func(b []byte) uint64
	action128 func(b []byte) [2]uint64
}

func BenchmarkDefault(b *testing.B) {
	all := []benchTask{{
		name: "Target64", version: 64, action: func(b []byte) uint64 {
			return Hash(b)
		}}, {
		name: "Target128", version: 128, action128: func(b []byte) [2]uint64 {
			return Hash128(b)
		}},
	}

	benchLen0_16(b, all)
	benchLen17_128(b, all)
	benchLen129_240(b, all)
	benchLen241_1024(b, all)
	benchScalar(b, all)
	benchAVX2(b, all)
	benchSSE2(b, all)
}

func benchLen0_16(b *testing.B, benchTasks []benchTask) {
	for _, v := range benchTasks {
		b.Run("Len0_16/"+v.name, func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				for i := 0; i <= 16; i++ {
					input := dat[b.N : b.N+i]
					if v.version == 64 {
						a := v.action(input)
						runtime.KeepAlive(a)
					} else {
						a := v.action128(input)
						runtime.KeepAlive(a)
					}
				}
			}
		})
	}
}

func benchLen17_128(b *testing.B, benchTasks []benchTask) {
	for _, v := range benchTasks {
		b.Run("Len17_128/"+v.name, func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				for i := 17; i <= 128; i++ {
					input := dat[b.N : b.N+i]
					if v.version == 64 {
						a := v.action(input)
						runtime.KeepAlive(a)
					} else {
						a := v.action128(input)
						runtime.KeepAlive(a)
					}
				}
			}
		})
	}
}

func benchLen129_240(b *testing.B, benchTasks []benchTask) {
	for _, v := range benchTasks {
		b.Run("Len129_240/"+v.name, func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				for i := 129; i <= 240; i++ {
					input := dat[b.N : b.N+i]
					if v.version == 64 {
						a := v.action(input)
						runtime.KeepAlive(a)
					} else {
						a := v.action128(input)
						runtime.KeepAlive(a)
					}
				}
			}
		})
	}
}

func benchLen241_1024(b *testing.B, benchTasks []benchTask) {
	avx2, sse2 = false, false
	for _, v := range benchTasks {
		b.Run("Len241_1024/"+v.name, func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				for i := 241; i <= 1024; i++ {
					input := dat[b.N : b.N+i]
					if v.version == 64 {
						a := v.action(input)
						runtime.KeepAlive(a)
					} else {
						a := v.action128(input)
						runtime.KeepAlive(a)
					}
				}
			}
		})
	}
}

func benchScalar(b *testing.B, benchTasks []benchTask) {
	avx2, sse2 = false, false
	for _, v := range benchTasks {
		b.Run("Scalar/"+v.name, func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				input := dat[n:33554432]
				if v.version == 64 {
					a := v.action(input)
					runtime.KeepAlive(a)
				} else {
					a := v.action128(input)
					runtime.KeepAlive(a)
				}
			}
		})
	}
}

func benchAVX2(b *testing.B, benchTasks []benchTask) {
	avx2, sse2 = true, false
	for _, v := range benchTasks {
		b.Run("AVX2/"+v.name, func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				input := dat[n:33554432]
				if v.version == 64 {
					a := v.action(input)
					runtime.KeepAlive(a)
				} else {
					a := v.action128(input)
					runtime.KeepAlive(a)
				}
			}
		})
	}
}
func benchSSE2(b *testing.B, benchTasks []benchTask) {
	avx2, sse2 = false, true
	for _, v := range benchTasks {
		b.Run("SSE2/"+v.name, func(b *testing.B) {
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				input := dat[n:33554432]
				if v.version == 64 {
					a := v.action(input)
					runtime.KeepAlive(a)
				} else {
					a := v.action128(input)
					runtime.KeepAlive(a)
				}
			}
		})
	}
}
