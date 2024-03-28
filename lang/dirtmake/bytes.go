package dirtmake

import (
	"unsafe"
)

type slice struct {
	data unsafe.Pointer
	len  int
	cap  int
}

//go:linkname mallocgc runtime.mallocgc
func mallocgc(size uintptr, typ unsafe.Pointer, needzero bool) unsafe.Pointer

// Bytes allocates a byte slice but does not clean up the memory it references.
// Throw a fatal error instead of panic if cap is greater than runtime.maxAlloc.
// NOTE: MUST set any byte element before it's read.
func Bytes(len, cap int) (b []byte) {
	if len < 0 || len > cap {
		panic("dirtmake.Bytes: len out of range")
	}
	p := mallocgc(uintptr(cap), nil, false)
	sh := (*slice)(unsafe.Pointer(&b))
	sh.data = p
	sh.len = len
	sh.cap = cap
	return
}
