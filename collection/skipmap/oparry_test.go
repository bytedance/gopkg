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

package skipmap

import (
	"testing"
	"unsafe"

	"github.com/bytedance/gopkg/lang/fastrand"
)

type dummy struct {
	data optionalArray
}

func TestOpArray(t *testing.T) {
	n := new(dummy)
	n.data.extra = new([op2]unsafe.Pointer)

	var array [maxLevel]unsafe.Pointer
	for i := 0; i < maxLevel; i++ {
		value := unsafe.Pointer(&dummy{})
		array[i] = value
		n.data.store(i, value)
	}

	for i := 0; i < maxLevel; i++ {
		if array[i] != n.data.load(i) || array[i] != n.data.atomicLoad(i) {
			t.Fatal(i, array[i], n.data.load(i))
		}
	}

	for i := 0; i < 1000; i++ {
		r := int(fastrand.Uint32n(maxLevel))
		value := unsafe.Pointer(&dummy{})
		if i%100 == 0 {
			value = nil
		}
		array[r] = value
		if fastrand.Uint32n(2) == 0 {
			n.data.store(r, value)
		} else {
			n.data.atomicStore(r, value)
		}
	}

	for i := 0; i < maxLevel; i++ {
		if array[i] != n.data.load(i) || array[i] != n.data.atomicLoad(i) {
			t.Fatal(i, array[i], n.data.load(i))
		}
	}
}
