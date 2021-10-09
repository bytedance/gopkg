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

//go:build amd64 && !gccgo && !appengine
// +build amd64,!gccgo,!appengine

package lscq

import (
	"unsafe"
)

type uint128 [2]uint64

func compareAndSwapUint128(addr *uint128, old1, old2, new1, new2 uint64) (swapped bool)

func loadUint128(addr *uint128) (val uint128)

func loadSCQNodePointer(addr unsafe.Pointer) (val scqNodePointer)

func loadSCQNodeUint64(addr unsafe.Pointer) (val scqNodeUint64)

func atomicTestAndSetFirstBit(addr *uint64) (val uint64)

func atomicTestAndSetSecondBit(addr *uint64) (val uint64)

func resetNode(addr unsafe.Pointer)

//go:nosplit
func compareAndSwapSCQNodePointer(addr *scqNodePointer, old, new scqNodePointer) (swapped bool) {
	// Ref: src/runtime/atomic_pointer.go:sync_atomic_CompareAndSwapPointer
	if runtimeEnableWriteBarrier() {
		runtimeatomicwb(&addr.data, new.data)
	}
	return compareAndSwapUint128((*uint128)(runtimenoescape(unsafe.Pointer(addr))), old.flags, uint64(uintptr(old.data)), new.flags, uint64(uintptr(new.data)))
}

func compareAndSwapSCQNodeUint64(addr *scqNodeUint64, old, new scqNodeUint64) (swapped bool) {
	return compareAndSwapUint128((*uint128)(unsafe.Pointer(addr)), old.flags, old.data, new.flags, new.data)
}

func runtimeEnableWriteBarrier() bool

//go:linkname runtimeatomicwb runtime.atomicwb
//go:noescape
func runtimeatomicwb(ptr *unsafe.Pointer, new unsafe.Pointer)

//go:linkname runtimenoescape runtime.noescape
func runtimenoescape(p unsafe.Pointer) unsafe.Pointer

//go:nosplit
func atomicWriteBarrier(ptr *unsafe.Pointer) {
	// For SCQ dequeue only. (fastpath)
	if runtimeEnableWriteBarrier() {
		runtimeatomicwb(ptr, nil)
	}
}
