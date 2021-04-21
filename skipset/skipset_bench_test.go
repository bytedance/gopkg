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

package skipset

import (
	"math"
	"strconv"
	"sync"
	"testing"
)

const initsize = 1 << 10 // for `contains` `1Remove9Add90Contains` `1Range9Remove90Add900Contains`
const randN = math.MaxUint32

func BenchmarkAdd(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewInt64()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Add(int64(fastrandn(randN)))
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Store(int64(fastrandn(randN)), nil)
			}
		})
	})
}

func BenchmarkContains100Hits(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewInt64()
		for i := 0; i < initsize; i++ {
			l.Add(int64(i))
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = l.Contains(int64(fastrandn(initsize)))
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
				_, _ = l.Load(int64(fastrandn(initsize)))
			}
		})
	})
}

func BenchmarkContains50Hits(b *testing.B) {
	const rate = 2
	b.Run("skipset", func(b *testing.B) {
		l := NewInt64()
		for i := 0; i < initsize*rate; i++ {
			if fastrandn(rate) == 0 {
				l.Add(int64(i))
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = l.Contains(int64(fastrandn(initsize * rate)))
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize*rate; i++ {
			if fastrandn(rate) == 0 {
				l.Store(int64(i), nil)
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(int64(fastrandn(initsize * rate)))
			}
		})
	})
}

func BenchmarkContainsNoHits(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewInt64()
		invalid := make([]int64, 0, initsize)
		for i := 0; i < initsize*2; i++ {
			if i%2 == 0 {
				l.Add(int64(i))
			} else {
				invalid = append(invalid, int64(i))
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = l.Contains(invalid[fastrandn(uint32(len(invalid)))])
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
				_, _ = l.Load(invalid[fastrandn(uint32(len(invalid)))])
			}
		})
	})
}

func Benchmark50Add50Contains(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewInt64()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(10)
				if u < 5 {
					l.Add(int64(fastrandn(randN)))
				} else {
					l.Contains(int64(fastrandn(randN)))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(10)
				if u < 5 {
					l.Store(int64(fastrandn(randN)), nil)
				} else {
					l.Load(int64(fastrandn(randN)))
				}
			}
		})
	})
}

func Benchmark30Add70Contains(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewInt64()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(10)
				if u < 3 {
					l.Add(int64(fastrandn(randN)))
				} else {
					l.Contains(int64(fastrandn(randN)))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(10)
				if u < 3 {
					l.Store(int64(fastrandn(randN)), nil)
				} else {
					l.Load(int64(fastrandn(randN)))
				}
			}
		})
	})
}

func Benchmark1Remove9Add90Contains(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewInt64()
		for i := 0; i < initsize; i++ {
			l.Add(int64(i))
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(100)
				if u < 9 {
					l.Add(int64(fastrandn(randN)))
				} else if u == 10 {
					l.Remove(int64(fastrandn(randN)))
				} else {
					l.Contains(int64(fastrandn(randN)))
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
				u := fastrandn(100)
				if u < 9 {
					l.Store(int64(fastrandn(randN)), nil)
				} else if u == 10 {
					l.Delete(int64(fastrandn(randN)))
				} else {
					l.Load(int64(fastrandn(randN)))
				}
			}
		})
	})
}

func Benchmark1Range9Remove90Add900Contains(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewInt64()
		for i := 0; i < initsize; i++ {
			l.Add(int64(i))
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(1000)
				if u == 0 {
					l.Range(func(score int64) bool {
						return true
					})
				} else if u > 10 && u < 20 {
					l.Remove(int64(fastrandn(randN)))
				} else if u >= 100 && u < 190 {
					l.Add(int64(fastrandn(randN)))
				} else {
					l.Contains(int64(fastrandn(randN)))
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
				u := fastrandn(1000)
				if u == 0 {
					l.Range(func(key, value interface{}) bool {
						return true
					})
				} else if u > 10 && u < 20 {
					l.Delete(fastrandn(randN))
				} else if u >= 100 && u < 190 {
					l.Store(fastrandn(randN), nil)
				} else {
					l.Load(fastrandn(randN))
				}
			}
		})
	})
}

func BenchmarkStringAdd(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewString()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Add(strconv.Itoa(int(fastrand())))
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Store(strconv.Itoa(int(fastrand())), nil)
			}
		})
	})
}

func BenchmarkStringContains50Hits(b *testing.B) {
	const rate = 2
	b.Run("skipset", func(b *testing.B) {
		l := NewString()
		for i := 0; i < initsize*rate; i++ {
			if fastrandn(rate) == 0 {
				l.Add(strconv.Itoa(i))
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = l.Contains(strconv.Itoa(int(fastrandn(initsize * rate))))
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		for i := 0; i < initsize*rate; i++ {
			if fastrandn(rate) == 0 {
				l.Store(strconv.Itoa(i), nil)
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = l.Load(strconv.Itoa(int(fastrandn(initsize * rate))))
			}
		})
	})
}

func BenchmarkString30Add70Contains(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewString()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(10)
				if u < 3 {
					l.Add(strconv.Itoa(int(fastrandn(randN))))
				} else {
					l.Contains(strconv.Itoa(int(fastrandn(randN))))
				}
			}
		})
	})
	b.Run("sync.Map", func(b *testing.B) {
		var l sync.Map
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(10)
				if u < 3 {
					l.Store(strconv.Itoa(int(fastrandn(randN))), nil)
				} else {
					l.Load(strconv.Itoa(int(fastrandn(randN))))
				}
			}
		})
	})
}

func BenchmarkString1Remove9Add90Contains(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewString()
		for i := 0; i < initsize; i++ {
			l.Add(strconv.Itoa(i))
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(100)
				if u < 9 {
					l.Add(strconv.Itoa(int(fastrandn(randN))))
				} else if u == 10 {
					l.Remove(strconv.Itoa(int(fastrandn(randN))))
				} else {
					l.Contains(strconv.Itoa(int(fastrandn(randN))))
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
				u := fastrandn(100)
				if u < 9 {
					l.Store(strconv.Itoa(int(fastrandn(randN))), nil)
				} else if u == 10 {
					l.Delete(strconv.Itoa(int(fastrandn(randN))))
				} else {
					l.Load(strconv.Itoa(int(fastrandn(randN))))
				}
			}
		})
	})
}

func BenchmarkString1Range9Remove90Add900Contains(b *testing.B) {
	b.Run("skipset", func(b *testing.B) {
		l := NewString()
		for i := 0; i < initsize; i++ {
			l.Add(strconv.Itoa(i))
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				u := fastrandn(1000)
				if u == 0 {
					l.Range(func(score string) bool {
						return true
					})
				} else if u > 10 && u < 20 {
					l.Remove(strconv.Itoa(int(fastrandn(randN))))
				} else if u >= 100 && u < 190 {
					l.Add(strconv.Itoa(int(fastrandn(randN))))
				} else {
					l.Contains(strconv.Itoa(int(fastrandn(randN))))
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
				u := fastrandn(1000)
				if u == 0 {
					l.Range(func(key, value interface{}) bool {
						return true
					})
				} else if u > 10 && u < 20 {
					l.Delete(strconv.Itoa(int(fastrandn(randN))))
				} else if u >= 100 && u < 190 {
					l.Store(strconv.Itoa(int(fastrandn(randN))), nil)
				} else {
					l.Load(strconv.Itoa(int(fastrandn(randN))))
				}
			}
		})
	})
}
