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
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/stretchr/testify/assert"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString(prefix string) string {
	b := make([]rune, 8)
	for i := range b {
		b[i] = letterRunes[fastrand.Intn(len(letterRunes))]
	}
	return prefix + string(b)
}

func TestFloat64Set(t *testing.T) {
	z := NewFloat64()
	assert.Zero(t, z.Len())
}

func TestFloat64SetAdd(t *testing.T) {
	z := NewFloat64()
	v := randString("")
	assert.True(t, z.Add(1, v))
	assert.False(t, z.Add(1, v))
}

func TestFloat64SetContains(t *testing.T) {
	z := NewFloat64()
	v := randString("")
	z.Add(1, v)
	assert.True(t, z.Contains(v))
	assert.False(t, z.Contains("no-such-"+v))
}

func TestFloat64SetScore(t *testing.T) {
	z := NewFloat64()
	v := randString("")
	s := rand.Float64()
	z.Add(s, v)
	as, ok := z.Score(v)
	assert.True(t, ok)
	assert.Equal(t, s, as)
	_, ok = z.Score("no-such-" + v)
	assert.False(t, ok)
}

func TestFloat64SetIncr(t *testing.T) {
	z := NewFloat64()
	_, ok := z.Score("t")
	assert.False(t, ok)

	// test first insert
	s, ok := z.IncrBy(1, "t")
	assert.False(t, ok)
	assert.Equal(t, 1.0, s)

	// test regular incr
	s, ok = z.IncrBy(2, "t")
	assert.True(t, ok)
	assert.Equal(t, 3.0, s)
}

func TestFloat64SetRemove(t *testing.T) {
	z := NewFloat64()
	// test first insert
	ok := z.Add(1, "t")
	assert.True(t, ok)
	_, ok = z.Remove("t")
	assert.True(t, ok)
}

func TestFloat64SetRank(t *testing.T) {
	z := NewFloat64()
	v := randString("")
	z.Add(1, v)
	// test rank of exist value
	assert.Equal(t, 0, z.Rank(v))
	// test rank of non-exist value
	assert.Equal(t, -1, z.Rank("no-such-"+v))
}

func TestFloat64SetRank_Many(t *testing.T) {
	const N = 1000
	z := NewFloat64()
	rand.Seed(time.Now().Unix())

	var vs []string
	for i := 0; i < N; i++ {
		v := randString("")
		z.Add(rand.Float64(), v)
		vs = append(vs, v)
	}
	for _, v := range vs {
		r := z.Rank(v)
		assert.NotEqual(t, -1, r)

		// verify rank by traversing level 0
		actualRank := 0
		x := z.list.header
		for x != nil {
			x = x.loadNext(0)
			if x.value == v {
				break
			}
			actualRank++
		}
		assert.Equal(t, v, x.value)
		assert.Equal(t, r, actualRank)
	}
}

func TestFloat64SetRank_UpdateScore(t *testing.T) {
	z := NewFloat64()
	rand.Seed(time.Now().Unix())

	var vs []string
	for i := 0; i < 100; i++ {
		v := fmt.Sprint(i)
		z.Add(rand.Float64(), v)
		vs = append(vs, v)
	}
	// Randomly update score
	for i := 0; i < 100; i++ {
		// 1/2
		if rand.Float64() > 0.5 {
			continue
		}
		z.Add(float64(i), fmt.Sprint(i))
	}

	for _, v := range vs {
		r := z.Rank(v)
		assert.NotEqual(t, -1, r)
		assert.Greater(t, z.Len(), r)

		// verify rank by traversing level 0
		actualRank := 0
		x := z.list.header
		for x != nil {
			x = x.loadNext(0)
			if x.value == v {
				break
			}
			actualRank++
		}
		assert.Equal(t, v, x.value)
		assert.Equal(t, r, actualRank)
	}
}

// Test whether the ramdom inserted values sorted
func TestFloat64SetIsSorted(t *testing.T) {
	const N = 1000
	z := NewFloat64()
	rand.Seed(time.Now().Unix())

	// Test whether the ramdom inserted values sorted
	for i := 0; i < N; i++ {
		z.Add(fastrand.Float64(), fmt.Sprint(i))
	}
	testIsSorted(t, z)
	testInternalSpan(t, z)

	// Randomly update score
	for i := 0; i < N; i++ {
		// 1/2
		if rand.Float64() > 0.5 {
			continue
		}
		z.Add(float64(i), fmt.Sprint(i))
	}

	testIsSorted(t, z)
	testInternalSpan(t, z)

	// Randomly add or delete value
	for i := 0; i < N; i++ {
		// 1/2
		if rand.Float64() > 0.5 {
			continue
		}
		z.Remove(fmt.Sprint(i))
	}
	testIsSorted(t, z)
	testInternalSpan(t, z)
}

func testIsSorted(t *testing.T, z *Float64Set) {
	var scores []float64
	for _, n := range z.Range(0, z.Len()-1) {
		scores = append(scores, n.Score)
	}
	assert.True(t, sort.Float64sAreSorted(scores))
}

func testInternalSpan(t *testing.T, z *Float64Set) {
	l := z.list
	for i := l.highestLevel - 1; i >= 0; i-- {
		x := l.header
		for x.loadNext(i) != nil {
			x = x.loadNext(i)
			span := x.loadSpan(i)
			from := x.value
			fromScore := x.score
			fromRank := l.Rank(fromScore, from)
			assert.NotEqual(t, -1, fromRank)

			if x.loadNext(i) != nil { // from -> to
				to := x.loadNext(i).value
				toScore := x.loadNext(i).score
				toRank := l.Rank(toScore, to)
				assert.NotEqual(t, -1, toRank)

				// span = to.rank - from.rank
				assert.Equalf(t, span, toRank-fromRank, "from %q (score: , rank: %d) to %q (score: %d, rank: %d), expect span: %d, actual: %d",
					from, fromScore, fromRank, to, toScore, toRank, span, toRank-fromRank)
			} else { // from -> nil
				// span = skiplist.len - from.rank
				assert.Equalf(t, l.length-fromRank, x.loadSpan(i), "%q (score: , rank: %d)", from, fromScore, fromRank)
			}
		}
	}
}

func TestFloat64SetRange(t *testing.T) {
	testFloat64SetRange(t, false)
}

func TestFloat64SetRevRange(t *testing.T) {
	testFloat64SetRange(t, true)
}

func testFloat64SetRange(t *testing.T, rev bool) {
	const N = 1000
	z := NewFloat64()
	for i := 0; i < N; i++ {
		z.Add(fastrand.Float64(), fmt.Sprint(i))
	}

	start, stop := func(a, b int) (int, int) {
		if a < b {
			return a, b
		} else {
			return b, a
		}
	}(fastrand.Intn(N), fastrand.Intn(N))
	var ns []Float64Node
	if rev {
		ns = z.RevRange(start, stop)
	} else {
		ns = z.Range(start, stop)
	}
	assert.Equal(t, stop-start+1, len(ns))
	for i, n := range ns {
		if rev {
			assert.Equal(t, z.Len()-1-(start+i), z.Rank(n.Value))
		} else {
			assert.Equal(t, start+i, z.Rank(n.Value))
		}
	}
}

func TestFloat64SetRange_Negative(t *testing.T) {
	const N = 1000
	z := NewFloat64()
	for i := 0; i < N; i++ {
		z.Add(fastrand.Float64(), fmt.Sprint(i))
	}
	ns := z.Range(-1, -1)
	assert.Equal(t, 1, len(ns))
	assert.Equal(t, z.Len()-1, z.Rank(ns[0].Value))
}

func TestFloat64SetRevRange_Negative(t *testing.T) {
	const N = 1000
	z := NewFloat64()
	for i := 0; i < N; i++ {
		z.Add(fastrand.Float64(), fmt.Sprint(i))
	}
	ns := z.RevRange(-1, -1)
	assert.Equal(t, 1, len(ns))
	assert.Equal(t, 0, z.Rank(ns[0].Value))
}

func TestFloat64SetRangeByScore(t *testing.T) {
	testFloat64SetRangeByScore(t, false)
}

func TestFloat64SetRangeByScoreWithOpt(t *testing.T) {
	z := NewFloat64()
	z.Add(1.0, "1")
	z.Add(1.1, "2")
	z.Add(2.0, "3")

	ns := z.RangeByScoreWithOpt(1.0, 2.0, RangeOpt{ExcludeMin: true})
	assert.Equal(t, 2, len(ns))
	assert.Equal(t, 1.1, ns[0].Score)
	assert.Equal(t, 2.0, ns[1].Score)

	ns = z.RangeByScoreWithOpt(1.0, 2.0, RangeOpt{ExcludeMin: true, ExcludeMax: true})
	assert.Equal(t, 1, len(ns))
	assert.Equal(t, 1.1, ns[0].Score)

	ns = z.RangeByScoreWithOpt(1.0, 2.0, RangeOpt{ExcludeMax: true})
	assert.Equal(t, 2, len(ns))
	assert.Equal(t, 1.0, ns[0].Score)
	assert.Equal(t, 1.1, ns[1].Score)

	ns = z.RangeByScoreWithOpt(2.0, 1.0, RangeOpt{})
	assert.Equal(t, 0, len(ns))
	ns = z.RangeByScoreWithOpt(2.0, 1.0, RangeOpt{ExcludeMin: true})
	assert.Equal(t, 0, len(ns))
	ns = z.RangeByScoreWithOpt(2.0, 1.0, RangeOpt{ExcludeMax: true})
	assert.Equal(t, 0, len(ns))

	ns = z.RangeByScoreWithOpt(1.0, 1.0, RangeOpt{ExcludeMax: true})
	assert.Equal(t, 0, len(ns))
	ns = z.RangeByScoreWithOpt(1.0, 1.0, RangeOpt{ExcludeMin: true})
	assert.Equal(t, 0, len(ns))
	ns = z.RangeByScoreWithOpt(1.0, 1.0, RangeOpt{})
	assert.Equal(t, 1, len(ns))
}

func TestFloat64SetRevRangeByScoreWithOpt(t *testing.T) {
	z := NewFloat64()
	z.Add(1.0, "1")
	z.Add(1.1, "2")
	z.Add(2.0, "3")

	ns := z.RevRangeByScoreWithOpt(2.0, 1.0, RangeOpt{ExcludeMax: true})
	assert.Equal(t, 2, len(ns))
	assert.Equal(t, 1.1, ns[0].Score)
	assert.Equal(t, 1.0, ns[1].Score)

	ns = z.RevRangeByScoreWithOpt(2.0, 1.0, RangeOpt{ExcludeMax: true, ExcludeMin: true})
	assert.Equal(t, 1, len(ns))
	assert.Equal(t, 1.1, ns[0].Score)

	ns = z.RevRangeByScoreWithOpt(2.0, 1.0, RangeOpt{ExcludeMin: true})
	assert.Equal(t, 2, len(ns))
	assert.Equal(t, 2.0, ns[0].Score)
	assert.Equal(t, 1.1, ns[1].Score)

	ns = z.RevRangeByScoreWithOpt(1.0, 2.0, RangeOpt{})
	assert.Equal(t, 0, len(ns))
	ns = z.RevRangeByScoreWithOpt(1.0, 2.0, RangeOpt{ExcludeMin: true})
	assert.Equal(t, 0, len(ns))
	ns = z.RevRangeByScoreWithOpt(1.0, 2.0, RangeOpt{ExcludeMax: true})
	assert.Equal(t, 0, len(ns))

	ns = z.RevRangeByScoreWithOpt(1.0, 1.0, RangeOpt{ExcludeMax: true})
	assert.Equal(t, 0, len(ns))
	ns = z.RevRangeByScoreWithOpt(1.0, 1.0, RangeOpt{ExcludeMin: true})
	assert.Equal(t, 0, len(ns))
	ns = z.RevRangeByScoreWithOpt(1.0, 1.0, RangeOpt{})
	assert.Equal(t, 1, len(ns))
}

func TestFloat64SetRevRangeByScore(t *testing.T) {
	testFloat64SetRangeByScore(t, true)
}

func testFloat64SetRangeByScore(t *testing.T, rev bool) {
	const N = 1000
	z := NewFloat64()
	for i := 0; i < N; i++ {
		z.Add(fastrand.Float64(), fmt.Sprint(i))
	}

	min, max := func(a, b float64) (float64, float64) {
		if a < b {
			return a, b
		} else {
			return b, a
		}
	}(fastrand.Float64(), fastrand.Float64())

	var ns []Float64Node
	if rev {
		ns = z.RevRangeByScore(max, min)
	} else {
		ns = z.RangeByScore(min, max)
	}
	var prev *float64
	for _, n := range ns {
		assert.LessOrEqual(t, min, n.Score)
		assert.GreaterOrEqual(t, max, n.Score)
		if prev != nil {
			if rev {
				assert.True(t, *prev >= n.Score)
			} else {
				assert.True(t, *prev <= n.Score)
			}
		}
		prev = &n.Score
	}
}

func TestFloat64SetCountWithOpt(t *testing.T) {
	testFloat64SetCountWithOpt(t, RangeOpt{})
	testFloat64SetCountWithOpt(t, RangeOpt{true, true})
	testFloat64SetCountWithOpt(t, RangeOpt{true, false})
	testFloat64SetCountWithOpt(t, RangeOpt{false, true})
}

func testFloat64SetCountWithOpt(t *testing.T, opt RangeOpt) {
	const N = 1000
	z := NewFloat64()
	for i := 0; i < N; i++ {
		z.Add(fastrand.Float64(), fmt.Sprint(i))
	}

	min, max := func(a, b float64) (float64, float64) {
		if a < b {
			return a, b
		} else {
			return b, a
		}
	}(fastrand.Float64(), fastrand.Float64())

	n := z.CountWithOpt(min, max, opt)
	actualN := 0
	for _, n := range z.Range(0, -1) {
		if opt.ExcludeMin {
			if n.Score <= min {
				continue
			}
		} else {
			if n.Score < min {
				continue
			}
		}
		if opt.ExcludeMax {
			if n.Score >= max {
				continue
			}
		} else {
			if n.Score > max {
				continue
			}
		}
		actualN++
	}
	assert.Equal(t, actualN, n)
}

func TestFloat64SetRemoveRangeByRank(t *testing.T) {
	const N = 1000
	z := NewFloat64()
	for i := 0; i < N; i++ {
		z.Add(fastrand.Float64(), fmt.Sprint(i))
	}

	start, stop := func(a, b int) (int, int) {
		if a < b {
			return a, b
		} else {
			return b, a
		}
	}(fastrand.Intn(N), fastrand.Intn(N))

	expectNs := z.Range(start, stop)
	actualNs := z.RemoveRangeByRank(start, stop)
	assert.Equal(t, expectNs, actualNs)

	// test whether removed
	for _, n := range actualNs {
		assert.False(t, z.Contains(n.Value))
	}
	assert.Equal(t, N, z.Len()+len(actualNs))
}

func TestFloat64SetRemoveRangeByRankConcurrently(t *testing.T) {
	const N = 10000
	z := NewFloat64()
	for i := 0; i < N; i++ {
		z.Add(float64(i), strconv.Itoa(i))
	}
	const G = 10
	wg := sync.WaitGroup{}
	for i := 0; i < G; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := fastrand.Intn(N / 2)
			stop := N/2 + start
			z.RemoveRangeByRank(start, stop)
		}()
	}
	wg.Wait()
}

func TestFloat64SetRemoveRangeByScore(t *testing.T) {
	const N = 1000
	z := NewFloat64()
	for i := 0; i < N; i++ {
		z.Add(fastrand.Float64(), fmt.Sprint(i))
	}

	min, max := func(a, b float64) (float64, float64) {
		if a < b {
			return a, b
		} else {
			return b, a
		}
	}(fastrand.Float64(), fastrand.Float64())

	expectNs := z.RangeByScore(min, max)
	actualNs := z.RemoveRangeByScore(min, max)
	assert.Equal(t, expectNs, actualNs)

	// test whether removed
	for _, n := range actualNs {
		assert.False(t, z.Contains(n.Value))
	}
	assert.Equal(t, N, z.Len()+len(actualNs))
}

func TestFloat64SetRemoveRangeByScoreWithOpt(t *testing.T) {
	testFloat64SetRemoveRangeByScoreWithOpt(t, RangeOpt{})
	testFloat64SetRemoveRangeByScoreWithOpt(t, RangeOpt{true, true})
	testFloat64SetRemoveRangeByScoreWithOpt(t, RangeOpt{true, false})
	testFloat64SetRemoveRangeByScoreWithOpt(t, RangeOpt{false, false})
}

func testFloat64SetRemoveRangeByScoreWithOpt(t *testing.T, opt RangeOpt) {
	const N = 1000
	z := NewFloat64()
	for i := 0; i < N; i++ {
		z.Add(fastrand.Float64(), fmt.Sprint(i))
	}

	min, max := func(a, b float64) (float64, float64) {
		if a < b {
			return a, b
		} else {
			return b, a
		}
	}(fastrand.Float64(), fastrand.Float64())

	expectNs := z.RangeByScoreWithOpt(min, max, opt)
	actualNs := z.RemoveRangeByScoreWithOpt(min, max, opt)
	assert.Equal(t, expectNs, actualNs)

	// test whether removed
	for _, n := range actualNs {
		assert.False(t, z.Contains(n.Value))
	}
	assert.Equal(t, N, z.Len()+len(actualNs))
}

func TestUnionFloat64(t *testing.T) {
	var zs []*Float64Set
	for i := 0; i < 10; i++ {
		z := NewFloat64()
		for j := 0; j < 100; j++ {
			if fastrand.Float64() > 0.8 {
				z.Add(fastrand.Float64(), fmt.Sprint(i))
			}
		}
		zs = append(zs, z)
	}
	z := UnionFloat64(zs...)
	for _, n := range z.Range(0, z.Len()-1) {
		var expectScore float64
		for i := 0; i < 10; i++ {
			s, _ := zs[i].Score(n.Value)
			expectScore += s
		}
		assert.Equal(t, expectScore, n.Score)
	}
}

func TestUnionFloat64_Empty(t *testing.T) {
	z := UnionFloat64()
	assert.Zero(t, z.Len())
}

func TestInterFloat64(t *testing.T) {
	var zs []*Float64Set
	for i := 0; i < 10; i++ {
		z := NewFloat64()
		for j := 0; j < 10; j++ {
			if fastrand.Float64() > 0.8 {
				z.Add(fastrand.Float64(), fmt.Sprint(i))
			}
		}
		zs = append(zs, z)
	}
	z := InterFloat64(zs...)
	for _, n := range z.Range(0, z.Len()-1) {
		var expectScore float64
		for i := 0; i < 10; i++ {
			s, ok := zs[i].Score(n.Value)
			assert.True(t, ok)
			expectScore += s
		}
		assert.Equal(t, expectScore, n.Score)
	}
}

func TestInterFloat64_Empty(t *testing.T) {
	z := InterFloat64()
	assert.Zero(t, z.Len())
}

func TestInterFloat64_Simple(t *testing.T) {
	z1 := NewFloat64()
	z1.Add(0, "1")
	z2 := NewFloat64()
	z2.Add(0, "1")
	z3 := NewFloat64()
	z3.Add(0, "2")

	z := InterFloat64(z1, z2, z3)
	assert.Zero(t, z.Len())
}
