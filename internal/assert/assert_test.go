// Copyright 2025 ByteDance Inc.
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

package assert

import "testing"

type nilTestStruct struct{}

func TestEqual(t *testing.T) {
	Equal(t, []int{1, 2}, []int{1, 2})
}

func TestNotEqual(t *testing.T) {
	NotEqual(t, []int{1, 2}, []int{2, 1})
}

func TestTrue(t *testing.T) {
	True(t, true)
}

func TestFalse(t *testing.T) {
	False(t, false)
}

func TestNil(t *testing.T) {
	var p *nilTestStruct
	Nil(t, p)

	var m map[string]int
	Nil(t, m)

	var s []int
	Nil(t, s)

	var i interface{} = p
	Nil(t, i)
}
