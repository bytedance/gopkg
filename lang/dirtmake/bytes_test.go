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

package dirtmake

import (
	"bytes"
	"fmt"
	"testing"
)

var data []byte

const block1kb = 1024

func BenchmarkParallelDirtBytes(b *testing.B) {
	src := make([]byte, block1kb)
	for i := range src {
		src[i] = byte(i)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			bs := Bytes(block1kb, block1kb)
			copy(bs, src)
			if !bytes.Equal(bs, src) {
				b.Fatalf("bytes not equal")
			}
		}
	})
}

func BenchmarkDirtBytes(b *testing.B) {
	for size := block1kb; size < block1kb*20; size += block1kb * 2 {
		b.Run(fmt.Sprintf("size=%dkb", size/block1kb), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				data = Bytes(size, size)
			}
		})
	}
}

func BenchmarkOriginBytes(b *testing.B) {
	for size := block1kb; size < block1kb*20; size += block1kb * 2 {
		b.Run(fmt.Sprintf("size=%dkb", size/block1kb), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				data = make([]byte, size)
			}
		})
	}
}
