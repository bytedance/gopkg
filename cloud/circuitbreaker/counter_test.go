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

package circuitbreaker

import (
	"sync"
	"testing"
)

func BenchmarkAtomicCounter_Add(b *testing.B) {
	c := atomicCounter{}
	b.SetParallelism(1000)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			c.Add(1)
		}
	})
}

func BenchmarkPerPCounter_Add(b *testing.B) {
	c := newPerPCounter()
	b.SetParallelism(1000)
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			c.Add(1)
		}
	})
}

func TestPerPCounter(t *testing.T) {
	numPerG := 1000
	numG := 1000
	c := newPerPCounter()
	c1 := atomicCounter{}
	var wg sync.WaitGroup
	wg.Add(numG)
	for i := 0; i < numG; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < numPerG; i++ {
				c.Add(1)
				c1.Add(1)
			}
		}()
	}
	wg.Wait()
	total := c.Get()
	total1 := c1.Get()
	if total != c1.Get() {
		t.Errorf("expected %d, get %d", total1, total)
	}
	c.Zero()
	c1.Zero()
	if c.Get() != 0 || c1.Get() != 0 {
		t.Errorf("zero failed")
	}
}
