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
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/bytedance/gopkg/lang/fastrand"
)

func Example() {
	l := NewInt()

	for _, v := range []int{10, 12, 15} {
		if l.Add(v) {
			fmt.Println("skipset add", v)
		}
	}

	if l.Contains(10) {
		fmt.Println("skipset contains 10")
	}

	l.Range(func(value int) bool {
		fmt.Println("skipset range found ", value)
		return true
	})

	l.Remove(15)
	fmt.Printf("skipset contains %d items\r\n", l.Len())
}

func TestIntSet(t *testing.T) {
	// Correctness.
	l := NewInt()
	if l.length != 0 {
		t.Fatal("invalid length")
	}
	if l.Contains(0) {
		t.Fatal("invalid contains")
	}

	if !l.Add(0) || l.length != 1 {
		t.Fatal("invalid add")
	}
	if !l.Contains(0) {
		t.Fatal("invalid contains")
	}
	if !l.Remove(0) || l.length != 0 {
		t.Fatal("invalid remove")
	}

	if !l.Add(20) || l.length != 1 {
		t.Fatal("invalid add")
	}
	if !l.Add(22) || l.length != 2 {
		t.Fatal("invalid add")
	}
	if !l.Add(21) || l.length != 3 {
		t.Fatal("invalid add")
	}

	var i int
	l.Range(func(score int) bool {
		if i == 0 && score != 20 {
			t.Fatal("invalid range")
		}
		if i == 1 && score != 21 {
			t.Fatal("invalid range")
		}
		if i == 2 && score != 22 {
			t.Fatal("invalid range")
		}
		i++
		return true
	})

	if !l.Remove(21) || l.length != 2 {
		t.Fatal("invalid remove")
	}

	i = 0
	l.Range(func(score int) bool {
		if i == 0 && score != 20 {
			t.Fatal("invalid range")
		}
		if i == 1 && score != 22 {
			t.Fatal("invalid range")
		}
		i++
		return true
	})

	const num = math.MaxInt16
	// Make rand shuffle array.
	// The testArray contains [1,num]
	testArray := make([]int, num)
	testArray[0] = num + 1
	for i := 1; i < num; i++ {
		// We left 0, because it is the default score for head and tail.
		// If we check the skipset contains 0, there must be something wrong.
		testArray[i] = int(i)
	}
	for i := len(testArray) - 1; i > 0; i-- { // Fisher–Yates shuffle
		j := fastrand.Uint32n(uint32(i + 1))
		testArray[i], testArray[j] = testArray[j], testArray[i]
	}

	// Concurrent add.
	var wg sync.WaitGroup
	for i := 0; i < num; i++ {
		i := i
		wg.Add(1)
		go func() {
			l.Add(testArray[i])
			wg.Done()
		}()
	}
	wg.Wait()
	if l.length != int64(num) {
		t.Fatalf("invalid length expected %d, got %d", num, l.length)
	}

	// Don't contains 0 after concurrent addion.
	if l.Contains(0) {
		t.Fatal("contains 0 after concurrent addion")
	}

	// Concurrent contains.
	for i := 0; i < num; i++ {
		i := i
		wg.Add(1)
		go func() {
			if !l.Contains(testArray[i]) {
				wg.Done()
				panic(fmt.Sprintf("add doesn't contains %d", i))
			}
			wg.Done()
		}()
	}
	wg.Wait()

	// Concurrent remove.
	for i := 0; i < num; i++ {
		i := i
		wg.Add(1)
		go func() {
			if !l.Remove(testArray[i]) {
				wg.Done()
				panic(fmt.Sprintf("can't remove %d", i))
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if l.length != 0 {
		t.Fatalf("invalid length expected %d, got %d", 0, l.length)
	}

	// Test all methods.
	const smallRndN = 1 << 8
	for i := 0; i < 1<<12; i++ {
		wg.Add(1)
		go func() {
			r := fastrand.Uint32n(num)
			if r < 333 {
				l.Add(int(fastrand.Uint32n(smallRndN)) + 1)
			} else if r < 666 {
				l.Contains(int(fastrand.Uint32n(smallRndN)) + 1)
			} else if r != 999 {
				l.Remove(int(fastrand.Uint32n(smallRndN)) + 1)
			} else {
				var pre int
				l.Range(func(score int) bool {
					if score <= pre { // 0 is the default value for header and tail score
						panic("invalid content")
					}
					pre = score
					return true
				})
			}
			wg.Done()
		}()
	}
	wg.Wait()

	// Correctness 2.
	var (
		x     = NewInt()
		y     = NewInt()
		count = 10000
	)

	for i := 0; i < count; i++ {
		x.Add(i)
	}

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			x.Range(func(score int) bool {
				if x.Remove(score) {
					if !y.Add(score) {
						panic("invalid add")
					}
				}
				return true
			})
			wg.Done()
		}()
	}
	wg.Wait()
	if x.Len() != 0 || y.Len() != count {
		t.Fatal("invalid length")
	}

	// Concurrent Add and Remove in small zone.
	x = NewInt()
	var (
		addcount    uint64 = 0
		removecount uint64 = 0
	)

	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 1000; i++ {
				if fastrand.Uint32n(2) == 0 {
					if x.Remove(int(fastrand.Uint32n(10))) {
						atomic.AddUint64(&removecount, 1)
					}
				} else {
					if x.Add(int(fastrand.Uint32n(10))) {
						atomic.AddUint64(&addcount, 1)
					}
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
	if addcount < removecount {
		panic("invalid count")
	}
	if addcount-removecount != uint64(x.Len()) {
		panic("invalid count")
	}

	pre := -1
	x.Range(func(score int) bool {
		if score <= pre {
			panic("invalid content")
		}
		pre = score
		return true
	})
}

func TestIntSetDesc(t *testing.T) {
	s := NewIntDesc()
	nums := []int{-1, 0, 5, 12}
	for _, v := range nums {
		s.Add(v)
	}
	i := len(nums) - 1
	s.Range(func(value int) bool {
		if nums[i] != value {
			t.Fatal("error")
		}
		i--
		return true
	})
}

func TestStringSet(t *testing.T) {
	x := NewString()
	if !x.Add("111") || x.Len() != 1 {
		t.Fatal("invalid")
	}
	if !x.Add("222") || x.Len() != 2 {
		t.Fatal("invalid")
	}
	if x.Add("111") || x.Len() != 2 {
		t.Fatal("invalid")
	}
	if !x.Contains("111") || !x.Contains("222") {
		t.Fatal("invalid")
	}
	if !x.Remove("111") || x.Len() != 1 {
		t.Fatal("invalid")
	}
	if !x.Remove("222") || x.Len() != 0 {
		t.Fatal("invalid")
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		i := i
		go func() {
			if !x.Add(strconv.Itoa(i)) {
				panic("invalid")
			}
			wg.Done()
		}()
	}
	wg.Wait()

	tmp := make([]int, 0, 100)
	x.Range(func(val string) bool {
		res, _ := strconv.Atoi(val)
		tmp = append(tmp, res)
		return true
	})
	sort.Ints(tmp)
	for i := 0; i < 100; i++ {
		if i != tmp[i] {
			t.Fatal("invalid")
		}
	}
}
