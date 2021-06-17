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

package lscq

import (
	"sync/atomic"
	"testing"

	"github.com/bytedance/gopkg/internal/benchmark/linkedq"
	"github.com/bytedance/gopkg/internal/benchmark/msq"
	"github.com/bytedance/gopkg/lang/fastrand"
)

type uint64queue interface {
	Enqueue(uint64) bool
	Dequeue() (uint64, bool)
}

type benchTask struct {
	name string
	New  func() uint64queue
}

type faa int64

func (data *faa) Enqueue(_ uint64) bool {
	atomic.AddInt64((*int64)(data), 1)
	return true
}

func (data *faa) Dequeue() (uint64, bool) {
	atomic.AddInt64((*int64)(data), -1)
	return 0, false
}

func BenchmarkDefault(b *testing.B) {
	all := []benchTask{{
		name: "LSCQ", New: func() uint64queue {
			return NewUint64()
		}}}
	all = append(all, benchTask{
		name: "LinkedQueue",
		New: func() uint64queue {
			return linkedq.New()
		},
	})
	all = append(all, benchTask{
		name: "MSQueue",
		New: func() uint64queue {
			return msq.New()
		},
	})
	// all = append(all, benchTask{
	// 	name: "FAA",
	// 	New: func() uint64queue {
	// 		return new(faa)
	// 	},
	// })
	// all = append(all, benchTask{
	// 	name: "channel",
	// 	New: func() uint64queue {
	// 		return newChannelQ(scqsize)
	// 	},
	// })
	benchEnqueueOnly(b, all)
	benchDequeueOnlyEmpty(b, all)
	benchPair(b, all)
	bench50Enqueue50Dequeue(b, all)
	bench30Enqueue70Dequeue(b, all)
	bench70Enqueue30Dequeue(b, all)
}

func reportalloc(b *testing.B) {
	// b.SetBytes(8)
	// b.ReportAllocs()
}

func benchPair(b *testing.B, benchTasks []benchTask) {
	for _, v := range benchTasks {
		b.Run("Pair/"+v.name, func(b *testing.B) {
			q := v.New()
			reportalloc(b)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					q.Enqueue(uint64(fastrand.Uint32()))
					q.Dequeue()
				}
			})
		})
	}
}

func bench50Enqueue50Dequeue(b *testing.B, benchTasks []benchTask) {
	for _, v := range benchTasks {
		b.Run("50Enqueue50Dequeue/"+v.name, func(b *testing.B) {
			q := v.New()
			b.ResetTimer()
			reportalloc(b)
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if fastrand.Uint32n(2) == 0 {
						q.Enqueue(uint64(fastrand.Uint32()))
					} else {
						q.Dequeue()
					}
				}
			})
		})
	}
}

func bench70Enqueue30Dequeue(b *testing.B, benchTasks []benchTask) {
	for _, v := range benchTasks {
		b.Run("70Enqueue30Dequeue/"+v.name, func(b *testing.B) {
			q := v.New()
			reportalloc(b)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if fastrand.Uint32n(10) > 2 {
						q.Enqueue(uint64(fastrand.Uint32()))
					} else {
						q.Dequeue()
					}
				}
			})
		})
	}
}

func bench30Enqueue70Dequeue(b *testing.B, benchTasks []benchTask) {
	for _, v := range benchTasks {
		b.Run("30Enqueue70Dequeue/"+v.name, func(b *testing.B) {
			q := v.New()
			reportalloc(b)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if fastrand.Uint32n(10) <= 2 {
						q.Enqueue(uint64(fastrand.Uint32()))
					} else {
						q.Dequeue()
					}
				}
			})
		})
	}
}

func benchEnqueueOnly(b *testing.B, benchTasks []benchTask) {
	for _, v := range benchTasks {
		b.Run("EnqueueOnly/"+v.name, func(b *testing.B) {
			q := v.New()
			reportalloc(b)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					q.Enqueue(uint64(fastrand.Uint32()))
				}
			})
		})
	}
}

func benchDequeueOnlyEmpty(b *testing.B, benchTasks []benchTask) {
	for _, v := range benchTasks {
		b.Run("DequeueOnlyEmpty/"+v.name, func(b *testing.B) {
			q := v.New()
			reportalloc(b)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					q.Dequeue()
				}
			})
		})
	}
}
