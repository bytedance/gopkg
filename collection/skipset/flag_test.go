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
	"testing"
)

func TestFlag(t *testing.T) {
	// Correctness.
	const (
		f0 = 1 << iota
		f1
		f2
		f3
		f4
		f5
		f6
		f7
	)
	x := &bitflag{}

	x.SetTrue(f1 | f3)
	if x.Get(f0) || !x.Get(f1) || x.Get(f2) || !x.Get(f3) || !x.MGet(f0|f1|f2|f3, f1|f3) {
		t.Fatal("invalid")
	}
	x.SetTrue(f1)
	x.SetTrue(f1 | f3)
	if x.data != f1+f3 {
		t.Fatal("invalid")
	}

	x.SetFalse(f1 | f2)
	if x.Get(f0) || x.Get(f1) || x.Get(f2) || !x.Get(f3) || !x.MGet(f0|f1|f2|f3, f3) {
		t.Fatal("invalid")
	}
	x.SetFalse(f1 | f2)
	if x.data != f3 {
		t.Fatal("invalid")
	}
}
