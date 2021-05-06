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

type Int64Set map[int64]struct{}

// NewInt64 returns an empty int64 set
func NewInt64() Int64Set {
	return make(map[int64]struct{})
}

// NewInt64WithSize returns an empty int64 set initialized with specific size
func NewInt64WithSize(size int) Int64Set {
	return make(map[int64]struct{}, size)
}

// Add adds the specified element to this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Int64Set) Add(value int64) bool {
	s[value] = struct{}{}
	return true
}

// Contains returns true if this set contains the specified element
func (s Int64Set) Contains(value int64) bool {
	if _, ok := s[value]; ok {
		return true
	}
	return false
}

// Remove removes the specified element from this set
// Always returns true due to the build-in map doesn't indicate caller whether the given element already exists
// Reserves the return type for future extension
func (s Int64Set) Remove(value int64) bool {
	delete(s, value)
	return true
}

// Range calls f sequentially for each value present in the hashset.
// If f returns false, range stops the iteration.
func (s Int64Set) Range(f func(value int64) bool) {
	for k := range s {
		if !f(k) {
			break
		}
	}
}

// Len returns the number of elements of this set
func (s Int64Set) Len() int {
	return len(s)
}
