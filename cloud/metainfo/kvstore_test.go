// Copyright 2023 ByteDance Inc.
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

package metainfo

import (
	"fmt"
	"testing"
)

func TestKVStore(t *testing.T) {
	store := newKVStore()
	store["a"] = "a"
	store["a"] = "b"
	if store["a"] != "b" {
		t.Fatal()
	}
	store.recycle()
	if store["a"] == "b" {
		t.Fatal()
	}
	store = newKVStore()
	if store["a"] == "b" {
		t.Fatal()
	}
}

func BenchmarkMap(b *testing.B) {
	for keys := 1; keys <= 1000; keys *= 10 {
		b.Run(fmt.Sprintf("keys=%d", keys), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m := make(map[string]string)
				for idx := 0; idx < 1000; idx++ {
					m[fmt.Sprintf("key-%d", idx)] = string('a' + byte(idx%26))
				}
			}
		})
	}
}

func BenchmarkKVStore(b *testing.B) {
	for keys := 1; keys <= 1000; keys *= 10 {
		b.Run(fmt.Sprintf("keys=%d", keys), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m := newKVStore()
				for idx := 0; idx < 1000; idx++ {
					m[fmt.Sprintf("key-%d", idx)] = string('a' + byte(idx%26))
				}
				m.recycle()
			}
		})
	}
}
