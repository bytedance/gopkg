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
	"math/rand"
	"testing"
)

type int64SetBool map[int64]bool

func newInt64Bool() *int64SetBool {
	return &int64SetBool{}
}

func (s *int64SetBool) Add(value int64) bool {
	(*s)[value] = true
	return true
}

func (s *int64SetBool) Contains(value int64) bool {
	if _, ok := (*s)[value]; ok {
		return true
	}
	return false
}

func (s *int64SetBool) Remove(value int64) bool {
	delete(*s, value)
	return true
}

func (s *int64SetBool) Range(f func(value int64) bool) {
	for k := range *s {
		if !f(k) {
			break
		}
	}
}

func (s *int64SetBool) Len() int {
	return len(*s)
}

type int64SetAdd map[int64]struct{}

func newInt64Add() *int64SetAdd {
	return &int64SetAdd{}
}

func (s *int64SetAdd) Add(value int64) bool {
	if s.Contains(value) {
		return true
	}
	(*s)[value] = struct{}{}
	return true
}

func (s *int64SetAdd) Contains(value int64) bool {
	if _, ok := (*s)[value]; ok {
		return true
	}
	return false
}

func (s *int64SetAdd) Remove(value int64) bool {
	if s.Contains(value) {
		delete(*s, value)
		return true
	}
	return false
}

func (s *int64SetAdd) Range(f func(value int64) bool) {
	for k := range *s {
		if !f(k) {
			break
		}
	}
}

func (s *int64SetAdd) Len() int {
	return len(*s)
}

const capacity = 10000000

var randomList [capacity]int64

func init() {
	for i := 0; i < capacity; i++ {
		randomList[i] = int64(rand.Int63())
	}
}

func BenchmarkValueAsBool(b *testing.B) {
	b.ResetTimer()
	l := newInt64Bool()
	for n := 0; n < b.N; n++ {
		l.Add(randomList[n%capacity])
	}
}

func BenchmarkValueAsEmptyStruct(b *testing.B) {
	b.ResetTimer()
	l := NewInt64()
	for n := 0; n < b.N; n++ {
		l.Add(randomList[n%capacity])
	}
}

func BenchmarkAddAfterContains(b *testing.B) {
	b.ResetTimer()
	l := newInt64Add()
	for n := 0; n < b.N; n++ {
		l.Add(randomList[n%capacity])
	}
}

func BenchmarkAddWithoutContains(b *testing.B) {
	b.ResetTimer()
	l := NewInt64()
	for n := 0; n < b.N; n++ {
		l.Add(randomList[n%capacity])
	}
}

func BenchmarkRemoveAfterContains_Missing(b *testing.B) {
	l := newInt64Add()
	for n := 0; n < b.N; n++ {
		l.Add(randomList[n%capacity])
	}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		l.Remove(int64(rand.Int63()))
	}
}

func BenchmarkRemoveWithoutContains_Missing(b *testing.B) {

	l := NewInt64()
	for n := 0; n < b.N; n++ {
		l.Add(randomList[n%capacity])
	}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		l.Remove(int64(rand.Int63()))
	}
}

func BenchmarkRemoveAfterContains_Hitting(b *testing.B) {
	l := newInt64Add()
	for n := 0; n < b.N; n++ {
		l.Add(randomList[n%capacity])
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		l.Remove(randomList[n%capacity])
	}
}

func BenchmarkRemoveWithoutContains_Hitting(b *testing.B) {
	l := NewInt64()
	for n := 0; n < b.N; n++ {
		l.Add(randomList[n%capacity])
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		l.Remove(randomList[n%capacity])
	}
}
