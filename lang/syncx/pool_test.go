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

package syncx

import (
	"sync"
	"testing"
)

func TestSyncXPool(t *testing.T) {
	var pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}
	for i := 0; i < 1024; i++ {
		b := pool.Get().([]byte)
		pool.Put(b)
	}

	var bs = make([][]byte, p)
	for i := 0; i < 1024; i++ {
		for i := 0; i < p; i++ {
			bs[i] = pool.Get().([]byte)
		}
		for i := 0; i < p; i++ {
			pool.Put(bs[i])
		}
	}
}

func TestRaceSyncXPool(t *testing.T) {
	var pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}
	parallel := 8
	var wg sync.WaitGroup
	wg.Add(parallel)
	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()
			var bs = make([][]byte, p)
			for i := 0; i < 8; i++ {
				for i := 0; i < p; i++ {
					bs[i] = pool.Get().([]byte)
				}
				for i := 0; i < p; i++ {
					pool.Put(bs[i])
				}
			}
		}()
	}
	wg.Wait()
}

var p = 1024

func BenchmarkSyncPool(b *testing.B) {
	var pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	var bs = make([][]byte, p)

	// benchmark
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < p; i++ {
			bs[i] = pool.Get().([]byte)
		}
		for i := 0; i < p; i++ {
			pool.Put(bs[i])
		}
	}
}

func BenchmarkSyncXPool(b *testing.B) {
	var pool = Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
		NoGC: true,
	}

	var bs = make([][]byte, p)

	// benchmark
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < p; i++ {
			bs[i] = pool.Get().([]byte)
		}
		for i := 0; i < p; i++ {
			pool.Put(bs[i])
		}
	}
}

func BenchmarkSyncPoolParallel(b *testing.B) {
	var pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	// benchmark
	b.ReportAllocs()
	b.SetParallelism(16)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var bs = make([][]byte, p)
		for pb.Next() {
			for i := 0; i < p; i++ {
				bs[i] = pool.Get().([]byte)
			}
			for i := 0; i < p; i++ {
				pool.Put(bs[i])
			}
		}
	})
}

func BenchmarkSyncXPoolParallel(b *testing.B) {
	var pool = Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
		NoGC: true,
	}

	// benchmark
	b.ReportAllocs()
	b.SetParallelism(16)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var bs = make([][]byte, p)
		for pb.Next() {
			for i := 0; i < p; i++ {
				bs[i] = pool.Get().([]byte)
			}
			for i := 0; i < p; i++ {
				pool.Put(bs[i])
			}
		}
	})
}

func BenchmarkSyncPoolParallel1(b *testing.B) {
	var pool = sync.Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
	}

	// benchmark
	b.ReportAllocs()
	b.SetParallelism(16)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var bs []byte
		for pb.Next() {
			bs = pool.Get().([]byte)
			pool.Put(bs)
		}
	})
}

func BenchmarkSyncXPoolParallel1(b *testing.B) {
	var pool = Pool{
		New: func() interface{} {
			return make([]byte, 1024)
		},
		NoGC: true,
	}

	// benchmark
	b.ReportAllocs()
	b.SetParallelism(16)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		var bs []byte
		for pb.Next() {
			bs = pool.Get().([]byte)
			pool.Put(bs)
		}
	})
}
