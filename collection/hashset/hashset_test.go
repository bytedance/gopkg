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

package hashset

import (
	"fmt"
	"testing"
)

func Example() {
	l := NewInt()

	for _, v := range []int{10, 12, 15} {
		l.Add(v)
	}

	if l.Contains(10) {
		fmt.Println("hashset contains 10")
	}

	l.Range(func(value int) bool {
		fmt.Println("hashset range found ", value)
		return true
	})

	l.Remove(15)
	fmt.Printf("hashset contains %d items\r\n", l.Len())
}

func TestIntSet(t *testing.T) {
	// Correctness.
	l := NewInt64()
	if l.Len() != 0 {
		t.Fatal("invalid length")
	}
	if l.Contains(0) {
		t.Fatal("invalid contains")
	}

	if l.Add(0); l.Len() != 1 {
		t.Fatal("invalid add")
	}
	if !l.Contains(0) {
		t.Fatal("invalid contains")
	}
	if l.Remove(0); l.Len() != 0 {
		t.Fatal("invalid remove")
	}

	if l.Add(20); l.Len() != 1 {
		t.Fatal("invalid add")
	}
	if l.Add(22); l.Len() != 2 {
		t.Fatal("invalid add")
	}
	if l.Add(21); l.Len() != 3 {
		t.Fatal("invalid add")
	}
	if l.Add(21); l.Len() != 3 {
		t.Fatal(l.Len(), " invalid add")
	}
}
