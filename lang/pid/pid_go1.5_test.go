// Copyright 2022 Cholerae Hu.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.

// This file may have been modified by ByteDance Inc. ("ByteDance Modifications"). All ByteDance Modifications are Copyright 2022 ByteDance Inc.

//go:build gc && go1.5
// +build gc,go1.5

package goid

import (
	"fmt"
	"testing"
	"unsafe"
)

var _ = unsafe.Sizeof(0)

//go:linkname procPin runtime.procPin
//go:nosplit
func procPin() int

//go:linkname procUnpin runtime.procUnpin
//go:nosplit
func procUnpin()

func getTID() int {
	tid := procPin()
	procUnpin()
	return tid
}

func TestParallelGetPid(t *testing.T) {
	ch := make(chan *string, 100)
	for i := 0; i < cap(ch); i++ {
		go func(i int) {
			id := GetPid()
			expected := getTID()
			if id == expected {
				ch <- nil
				return
			}
			s := fmt.Sprintf("Expected %d, but got %d", expected, id)
			ch <- &s
		}(i)
	}

	for i := 0; i < cap(ch); i++ {
		val := <-ch
		if val != nil {
			t.Fatal(*val)
		}
	}
}

func TestGetPid(t *testing.T) {
	p1 := GetPid()
	p2 := getTID()
	if p1 != p2 {
		t.Fatalf("The result of GetPid %d procPin %d are not equal!", p1, p2)
	}
}

func BenchmarkGetPid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetPid()
	}
}

func BenchmarkParallelGetPid(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			GetPid()
		}
	})
}
