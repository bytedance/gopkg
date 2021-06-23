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

package lscq

import (
	"unsafe"

	"golang.org/x/sys/cpu"
)

const (
	scqsize       = 1 << 16
	cacheLineSize = unsafe.Sizeof(cpu.CacheLinePad{})
)

func uint64Get63(value uint64) uint64 {
	return value & ((1 << 63) - 1)
}

func uint64Get1(value uint64) bool {
	return (value & (1 << 63)) == (1 << 63)
}

func uint64GetAll(value uint64) (bool, uint64) {
	return (value & (1 << 63)) == (1 << 63), value & ((1 << 63) - 1)
}

func loadSCQFlags(flags uint64) (isSafe bool, isEmpty bool, cycle uint64) {
	isSafe = (flags & (1 << 63)) == (1 << 63)
	isEmpty = (flags & (1 << 62)) == (1 << 62)
	cycle = flags & ((1 << 62) - 1)
	return isSafe, isEmpty, cycle
}

func newSCQFlags(isSafe bool, isEmpty bool, cycle uint64) uint64 {
	v := cycle & ((1 << 62) - 1)
	if isSafe {
		v += 1 << 63
	}
	if isEmpty {
		v += 1 << 62
	}
	return v
}

func cacheRemap16Byte(index uint64) uint64 {
	const cacheLineSize = cacheLineSize / 2
	rawIndex := index & uint64(scqsize-1)
	cacheLineNum := (rawIndex) % (scqsize / uint64(cacheLineSize))
	cacheLineIdx := rawIndex / (scqsize / uint64(cacheLineSize))
	return cacheLineNum*uint64(cacheLineSize) + cacheLineIdx
}
