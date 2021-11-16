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
	"fmt"
	"sync"
	"testing"

	"github.com/bytedance/gopkg/lang/fastrand"
)

func BenchmarkRWMutex100Read(b *testing.B) {
	for nWorker := 1; nWorker <= 256; nWorker <<= 2 {
		b.Run(fmt.Sprintf("syncx-%d", nWorker), func(b *testing.B) {
			benchmarkSyncxRWMutexNWriteNRead(b, nWorker, 0)
		})

		b.Run(fmt.Sprintf("sync-%d", nWorker), func(b *testing.B) {
			benchmarkSyncRWMutexNWriteNRead(b, nWorker, 0)
		})
	}
}

func BenchmarkRWMutex1Write99Read(b *testing.B) {
	for nWorker := 1; nWorker <= 256; nWorker <<= 2 {
		b.Run(fmt.Sprintf("syncx-%d", nWorker), func(b *testing.B) {
			benchmarkSyncxRWMutexNWriteNRead(b, nWorker, 1)
		})

		b.Run(fmt.Sprintf("sync-%d", nWorker), func(b *testing.B) {
			benchmarkSyncRWMutexNWriteNRead(b, nWorker, 1)
		})
	}
}

func BenchmarkRWMutex10Write90Read(b *testing.B) {
	for nWorker := 1; nWorker <= 256; nWorker <<= 2 {
		b.Run(fmt.Sprintf("syncx-%d", nWorker), func(b *testing.B) {
			benchmarkSyncxRWMutexNWriteNRead(b, nWorker, 10)
		})

		b.Run(fmt.Sprintf("sync-%d", nWorker), func(b *testing.B) {
			benchmarkSyncRWMutexNWriteNRead(b, nWorker, 10)
		})
	}
}

func BenchmarkRWMutex50Write50Read(b *testing.B) {
	for nWorker := 1; nWorker <= 256; nWorker <<= 2 {
		b.Run(fmt.Sprintf("syncx-%d", nWorker), func(b *testing.B) {
			benchmarkSyncxRWMutexNWriteNRead(b, nWorker, 50)
		})

		b.Run(fmt.Sprintf("sync-%d", nWorker), func(b *testing.B) {
			benchmarkSyncRWMutexNWriteNRead(b, nWorker, 50)
		})
	}
}

func benchmarkSyncRWMutexNWriteNRead(b *testing.B, nWorker, nWrite int) {
	var res int // A mock resource we contention for

	var mu sync.RWMutex
	mu.Lock()

	var wg sync.WaitGroup
	wg.Add(nWorker)

	n := b.N
	quota := n / nWorker

	for g := nWorker; g > 0; g-- {
		// Comuse remaining quota
		if g == 1 {
			quota = n
		}
		go func(quota int) {
			for i := 0; i < quota; i++ {
				if fastrand.Intn(100) > nWrite-1 {
					mu.RLock()
					_ = res
					mu.RUnlock()
				} else {
					mu.Lock()
					res++
					mu.Unlock()
				}
			}
			wg.Done()
		}(quota)

		n -= quota
	}

	if n != 0 {
		b.Fatalf("Incorrect quota assignments: %v remaining", n)
	}

	b.ResetTimer()
	mu.Unlock()
	wg.Wait()
}

func benchmarkSyncxRWMutexNWriteNRead(b *testing.B, nWorker, nWrite int) {
	var res int // A mock resource we contention for

	mu := NewRWMutex()
	mu.Lock()

	var wg sync.WaitGroup
	wg.Add(nWorker)

	n := b.N
	quota := n / nWorker

	for g := nWorker; g > 0; g-- {
		// Comuse remaining quota
		if g == 1 {
			quota = n
		}
		go func(quota int) {
			rmu := mu.RLocker()
			for i := 0; i < quota; i++ {
				if fastrand.Intn(100) > nWrite-1 {
					rmu.Lock()
					_ = res
					rmu.Unlock()
				} else {
					mu.Lock()
					res++
					mu.Unlock()
				}
			}
			wg.Done()
		}(quota)

		n -= quota
	}

	if n != 0 {
		b.Fatalf("Incorrect quota assignments: %v remaining", n)
	}

	b.ResetTimer()
	mu.Unlock()
	wg.Wait()
}
