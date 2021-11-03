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

package zset

import (
	"math"
	"strconv"
	"testing"

	"github.com/bytedance/gopkg/lang/fastrand"
)

const initsize = 1 << 10
const randN = math.MaxUint32

func BenchmarkContains100Hits(b *testing.B) {
	benchmarkContainsNHits(b, 100)
}

func BenchmarkContains50Hits(b *testing.B) {
	benchmarkContainsNHits(b, 50)
}

func BenchmarkContainsNoHits(b *testing.B) {
	benchmarkContainsNHits(b, 0)
}

func benchmarkContainsNHits(b *testing.B, n int) {
	b.Run("sortedset", func(b *testing.B) {
		z := NewFloat64()
		var vals []string
		for i := 0; i < initsize; i++ {
			val := strconv.Itoa(i)
			vals = append(vals, val)
			if fastrand.Intn(100)+1 <= n {
				z.Add(fastrand.Float64(), val)
			}
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = z.Contains(vals[fastrand.Intn(initsize)])
			}
		})
	})
}

func BenchmarkAdd(b *testing.B) {
	benchmarkNAddNIncrNRemoveNContains(b, 100, 0, 0, 0)
}

func Benchmark1Add99Contains(b *testing.B) {
	benchmarkNAddNIncrNRemoveNContains(b, 1, 0, 0, 99)
}

func Benchmark10Add90Contains(b *testing.B) {
	benchmarkNAddNIncrNRemoveNContains(b, 10, 0, 0, 90)
}

func Benchmark50Add50Contains(b *testing.B) {
	benchmarkNAddNIncrNRemoveNContains(b, 50, 0, 0, 50)
}

func Benchmark1Add3Incr6Remove90Contains(b *testing.B) {
	benchmarkNAddNIncrNRemoveNContains(b, 1, 3, 6, 90)
}

func benchmarkNAddNIncrNRemoveNContains(b *testing.B, nAdd, nIncr, nRemove, nContains int) {
	// var anAdd, anIncr, anRemove, anContains int

	b.Run("sortedset", func(b *testing.B) {
		z := NewFloat64()
		var vals []string
		var scores []float64
		var ops []int
		for i := 0; i < initsize; i++ {
			vals = append(vals, strconv.Itoa(fastrand.Intn(randN)))
			scores = append(scores, fastrand.Float64())
			ops = append(ops, fastrand.Intn(100))
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				r := fastrand.Intn(initsize)
				val := vals[r]
				if u := ops[r] + 1; u <= nAdd {
					// anAdd++
					z.Add(scores[r], val)
				} else if u-nAdd <= nIncr {
					// anIncr++
					z.IncrBy(scores[r], val)
				} else if u-nAdd-nIncr <= nRemove {
					// anRemove++
					z.Remove(val)
				} else if u-nAdd-nIncr-nRemove <= nContains {
					// anContains++
					z.Contains(val)
				}
			}
		})
		// b.Logf("N: %d, Add: %f, Incr: %f, Remove: %f, Contains: %f", b.N, float64(anAdd)/float64(b.N), float64(anIncr)/float64(b.N), float64(anRemove)/float64(b.N), float64(anContains)/float64(b.N))
	})
}
