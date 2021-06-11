// +build amd64,!gccgo,!appengine

package lscq

import "unsafe"

type uint128 [2]uint64

func compareAndSwapUint128(addr *uint128, old, new uint128) (swapped bool)

func loadUint128(addr *uint128) (val uint128)

func loadSCQNodePointer(addr unsafe.Pointer) (val scqNodePointer)

func loadSCQNodeUint64(addr unsafe.Pointer) (val scqNodeUint64)

func compareAndSwapSCQNodePointer(addr *scqNodePointer, old, new scqNodePointer) (swapped bool)

func compareAndSwapSCQNodeUint64(addr *scqNodeUint64, old, new scqNodeUint64) (swapped bool)

func atomicTestAndSetFirstBit(addr *uint64) (val uint64)

func atomicTestAndSetSecondBit(addr *uint64) (val uint64)
