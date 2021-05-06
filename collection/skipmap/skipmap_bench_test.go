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

package skipmap

import (
	"math"
	"strconv"
	"sync"
	"testing"

	"github.com/bytedance/gopkg/lang/fastrand"
)

const initsize = 1 << 10 // for `load` `1Delete9Store90Load` `1Range9Delete90Store900Load`
const randN = math.MaxUint32

func BenchmarkStore(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewInt64()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Store(int64(fastrand.Uint32n(randN)), nil)
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Store(int64(fastrand.Uint32n(randN)), nil)
			}
		})
	})
}

func BenchmarkLoad100Hits(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewInt64()
		for i := 0; i < initsize; i++ {
			l.Store(int64(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(int64(fastrand.Uint32n(initsize)))
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize; i++ {
			l.Store(int64(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(int64(fastrand.Uint32n(initsize)))
			}
		})
	})
}

func BenchmarkLoad50Hits(b *testing.B) {
	const rate = 2
	b.Run("skipmap", func(b *testing.B) {
		l := NewInt64()
		for i := 0; i < initsize*rate; i++ {
			if fastrand.Uint32n(rate) == 0 {
				l.Store(int64(i), nil)
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(int64(fastrand.Uint32n(initsize * rate)))
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize*rate; i++ {
			if fastrand.Uint32n(rate) == 0 {
				l.Store(int64(i), nil)
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(int64(fastrand.Uint32n(initsize * rate)))
			}
		})
	})
}

func BenchmarkLoadNoHits(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewInt64()
		invalid := make([]int64, 0, initsize)
		for i := 0; i < initsize*2; i++ {
			if i%2 == 0 {
				l.Store(int64(i), nil)
			} else {
				invalid = append(invalid, int64(i))
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(invalid[fastrand.Uint32n(uint32(len(invalid)))])
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		invalid := make([]int64, 0, initsize)
		for i := 0; i < initsize*2; i++ {
			if i%2 == 0 {
				l.Store(int64(i), nil)
			} else {
				invalid = append(invalid, int64(i))
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(invalid[fastrand.Uint32n(uint32(len(invalid)))])
			}
		})
	})
}

func Benchmark50Store50Load(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewInt64()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(10)
				if u < 5 {
					l.Store(int64(fastrand.Uint32n(randN)), nil)
				} else {
					l.Load(int64(fastrand.Uint32n(randN)))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(10)
				if u < 5 {
					l.Store(int64(fastrand.Uint32n(randN)), nil)
				} else {
					l.Load(int64(fastrand.Uint32n(randN)))
				}
			}
		})
	})
}

func Benchmark30Store70Load(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewInt64()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(10)
				if u < 3 {
					l.Store(int64(fastrand.Uint32n(randN)), nil)
				} else {
					l.Load(int64(fastrand.Uint32n(randN)))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(10)
				if u < 3 {
					l.Store(int64(fastrand.Uint32n(randN)), nil)
				} else {
					l.Load(int64(fastrand.Uint32n(randN)))
				}
			}
		})
	})
}

func Benchmark1Delete9Store90Load(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewInt64()
		for i := 0; i < initsize; i++ {
			l.Store(int64(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(100)
				if u < 9 {
					l.Store(int64(fastrand.Uint32n(randN)), nil)
				} else if u == 10 {
					l.Delete(int64(fastrand.Uint32n(randN)))
				} else {
					l.Load(int64(fastrand.Uint32n(randN)))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize; i++ {
			l.Store(int64(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(100)
				if u < 9 {
					l.Store(int64(fastrand.Uint32n(randN)), nil)
				} else if u == 10 {
					l.Delete(int64(fastrand.Uint32n(randN)))
				} else {
					l.Load(int64(fastrand.Uint32n(randN)))
				}
			}
		})
	})
}

func Benchmark1Range9Delete90Store900Load(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewInt64()
		for i := 0; i < initsize; i++ {
			l.Store(int64(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(1000)
				if u == 0 {
					l.Range(func(key int64, value interface{}) bool {
						return true
					})
				} else if u > 10 && u < 20 {
					l.Delete(int64(fastrand.Uint32n(randN)))
				} else if u >= 100 && u < 190 {
					l.Store(int64(fastrand.Uint32n(randN)), nil)
				} else {
					l.Load(int64(fastrand.Uint32n(randN)))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize; i++ {
			l.Store(int64(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(1000)
				if u == 0 {
					l.Range(func(key, value interface{}) bool {
						return true
					})
				} else if u > 10 && u < 20 {
					l.Delete(int64(fastrand.Uint32n(randN)))
				} else if u >= 100 && u < 190 {
					l.Store(int64(fastrand.Uint32n(randN)), nil)
				} else {
					l.Load(int64(fastrand.Uint32n(randN)))
				}
			}
		})
	})
}

func BenchmarkStringStore(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewString()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Store(strconv.Itoa(int(fastrand.Uint32())), nil)
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Store(strconv.Itoa(int(fastrand.Uint32())), nil)
			}
		})
	})
}

func BenchmarkStringLoad50Hits(b *testing.B) {
	const rate = 2
	b.Run("skipmap", func(b *testing.B) {
		l := NewString()
		for i := 0; i < initsize*rate; i++ {
			if fastrand.Uint32n(rate) == 0 {
				l.Store(strconv.Itoa(i), nil)
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(strconv.Itoa(int(fastrand.Uint32n(initsize * rate))))
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize*rate; i++ {
			if fastrand.Uint32n(rate) == 0 {
				l.Store(strconv.Itoa(i), nil)
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(strconv.Itoa(int(fastrand.Uint32n(initsize * rate))))
			}
		})
	})
}

func BenchmarkString30Store70Load(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewString()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(10)
				if u < 3 {
					l.Store(strconv.Itoa(int(fastrand.Uint32n(randN))), nil)
				} else {
					l.Load(strconv.Itoa(int(fastrand.Uint32n(randN))))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(10)
				if u < 3 {
					l.Store(strconv.Itoa(int(fastrand.Uint32n(randN))), nil)
				} else {
					l.Load(strconv.Itoa(int(fastrand.Uint32n(randN))))
				}
			}
		})
	})
}

func BenchmarkString1Delete9Store90Load(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewString()
		for i := 0; i < initsize; i++ {
			l.Store(strconv.Itoa(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(100)
				if u < 9 {
					l.Store(strconv.Itoa(int(fastrand.Uint32n(randN))), nil)
				} else if u == 10 {
					l.Delete(strconv.Itoa(int(fastrand.Uint32n(randN))))
				} else {
					l.Load(strconv.Itoa(int(fastrand.Uint32n(randN))))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize; i++ {
			l.Store(strconv.Itoa(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(100)
				if u < 9 {
					l.Store(strconv.Itoa(int(fastrand.Uint32n(randN))), nil)
				} else if u == 10 {
					l.Delete(strconv.Itoa(int(fastrand.Uint32n(randN))))
				} else {
					l.Load(strconv.Itoa(int(fastrand.Uint32n(randN))))
				}
			}
		})
	})
}

func BenchmarkString1Range9Delete90Store900Load(b *testing.B) {
	b.Run("skipmap", func(b *testing.B) {
		l := NewString()
		for i := 0; i < initsize; i++ {
			l.Store(strconv.Itoa(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(1000)
				if u == 0 {
					l.Range(func(key string, value interface{}) bool {
						return true
					})
				} else if u > 10 && u < 20 {
					l.Delete(strconv.Itoa(int(fastrand.Uint32n(randN))))
				} else if u >= 100 && u < 190 {
					l.Store(strconv.Itoa(int(fastrand.Uint32n(randN))), nil)
				} else {
					l.Load(strconv.Itoa(int(fastrand.Uint32n(randN))))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize; i++ {
			l.Store(strconv.Itoa(i), nil)
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrand.Uint32n(1000)
				if u == 0 {
					l.Range(func(key, value interface{}) bool {
						return true
					})
				} else if u > 10 && u < 20 {
					l.Delete(strconv.Itoa(int(fastrand.Uint32n(randN))))
				} else if u >= 100 && u < 190 {
					l.Store(strconv.Itoa(int(fastrand.Uint32n(randN))), nil)
				} else {
					l.Load(strconv.Itoa(int(fastrand.Uint32n(randN))))
				}
			}
		})
	})
}
