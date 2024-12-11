// Copyright 2024 ByteDance Inc.
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
	"testing"

	"github.com/bytedance/gopkg/lang/fastrand"
)

func BenchmarkLoadOrStoreExist(b *testing.B) {
	m := NewInt()
	m.Store(1, 1)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			m.LoadOrStore(1, 1)
		}
	})
}

func BenchmarkLoadOrStoreLazyExist(b *testing.B) {
	m := NewInt()
	m.Store(1, 1)
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			m.LoadOrStoreLazy(1, func() interface{} { return 1 })
		}
	})
}

func BenchmarkLoadOrStoreExistSingle(b *testing.B) {
	m := NewInt()
	m.Store(1, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.LoadOrStore(1, 1)
	}
}

func BenchmarkLoadOrStoreLazyExistSingle(b *testing.B) {
	m := NewInt()
	m.Store(1, 1)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.LoadOrStoreLazy(1, func() interface{} { return 1 })
	}
}

func BenchmarkLoadOrStoreRandom(b *testing.B) {
	m := NewInt()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			m.LoadOrStore(fastrand.Int(), 1)
		}
	})
}

func BenchmarkLoadOrStoreLazyRandom(b *testing.B) {
	m := NewInt()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			m.LoadOrStoreLazy(fastrand.Int(), func() interface{} { return 1 })
		}
	})
}

func BenchmarkLoadOrStoreRandomSingle(b *testing.B) {
	m := NewInt()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.LoadOrStore(fastrand.Int(), 1)
	}
}

func BenchmarkLoadOrStoreLazyRandomSingle(b *testing.B) {
	m := NewInt()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.LoadOrStoreLazy(fastrand.Int(), func() interface{} { return 1 })
	}
}
