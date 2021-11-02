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

package zset

import (
	"unsafe"
)

const (
	op1 = 4
	op2 = maxLevel - op1 // TODO: not sure that whether 4 is the best number for op1([28]Pointer for op2).
)

type listLevel struct {
	next unsafe.Pointer // the forward pointer
	span int            // span is count of level 0 element to next element in current level
}
type optionalArray struct {
	base  [op1]listLevel
	extra *([op2]listLevel)
}

func (a *optionalArray) init(level int) {
	if level > op1 {
		a.extra = new([op2]listLevel)
	}
}

func (a *optionalArray) loadNext(i int) unsafe.Pointer {
	if i < op1 {
		return a.base[i].next
	}
	return a.extra[i-op1].next
}

func (a *optionalArray) storeNext(i int, p unsafe.Pointer) {
	if i < op1 {
		a.base[i].next = p
		return
	}
	a.extra[i-op1].next = p
}

func (a *optionalArray) loadSpan(i int) int {
	if i < op1 {
		return a.base[i].span
	}
	return a.extra[i-op1].span
}

func (a *optionalArray) storeSpan(i int, s int) {
	if i < op1 {
		a.base[i].span = s
		return
	}
	a.extra[i-op1].span = s
}
