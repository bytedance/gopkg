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

// +build amd64

package timex

import "testing"

func BenchmarkFenceRDTSC(b *testing.B) {
	for i := 0; i < b.N; i++ {
		FenceRDTSC()
	}
}

func BenchmarkRDTSC(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RDTSC()
	}
}

func BenchmarkRDTSCP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RDTSCP()
	}
}

func TestRDTSC(t *testing.T) {
	v1 := FenceRDTSC()
	v2 := RDTSC()
	v3 := RDTSCP()
	t.Log(v1, v2, v3)
}
