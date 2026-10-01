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

// Package wyhash implements https://github.com/wangyi-fudan/wyhash
// (final version 4.3, little-endian flavor, WYHASH_CONDOM=1).
package wyhash

import (
	"math/bits"
	"unsafe"

	"github.com/bytedance/gopkg/internal/hack"
	"github.com/bytedance/gopkg/internal/runtimex"
)

// wyhash final version 4.3 default secret parameters.
const (
	s0 = 0x2d358dccaa6c78a5
	s1 = 0x8bb84b93962eacc9
	s2 = 0x4b33a62ed433d4a3
	s3 = 0x4d5a2da51de1aa47
)

// DefaultSeed is the seed used by Sum64 and Sum64String. It matches the
// common wyhash usage `wyhash(key, len, 0, _wyp)`.
const DefaultSeed = 0

func _wymix(a, b uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	return hi ^ lo
}

//go:nosplit
func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func Sum64(data []byte) uint64 {
	return Sum64WithSeed(data, DefaultSeed)
}

func Sum64String(data string) uint64 {
	return Sum64StringWithSeed(data, DefaultSeed)
}

func Sum64WithSeed(data []byte, seed uint64) uint64 {
	return Sum64StringWithSeed(hack.BytesToString(data), seed)
}

func Sum64StringWithSeed(data string, seed uint64) uint64 {
	paddr := *(*unsafe.Pointer)(unsafe.Pointer(&data))
	length := uint64(len(data))
	seed ^= _wymix(seed^s0, s1)

	var a, b uint64
	if length <= 16 {
		if length >= 4 {
			a = (uint64(runtimex.ReadUnaligned32(paddr)) << 32) | uint64(runtimex.ReadUnaligned32(add(paddr, uintptr((length>>3)<<2))))
			b = (uint64(runtimex.ReadUnaligned32(add(paddr, uintptr(length-4)))) << 32) | uint64(runtimex.ReadUnaligned32(add(paddr, uintptr(length-4-((length>>3)<<2)))))
		} else if length > 0 {
			a = uint64(*(*byte)(paddr))<<16 | uint64(*(*byte)(add(paddr, uintptr(length>>1))))<<8 | uint64(*(*byte)(add(paddr, uintptr(length-1))))
			b = 0
		} else {
			a, b = 0, 0
		}
	} else {
		i := length
		if i >= 48 {
			see1, see2 := seed, seed
			for {
				seed = _wymix(runtimex.ReadUnaligned64(paddr)^s1, runtimex.ReadUnaligned64(add(paddr, 8))^seed)
				see1 = _wymix(runtimex.ReadUnaligned64(add(paddr, 16))^s2, runtimex.ReadUnaligned64(add(paddr, 24))^see1)
				see2 = _wymix(runtimex.ReadUnaligned64(add(paddr, 32))^s3, runtimex.ReadUnaligned64(add(paddr, 40))^see2)
				paddr = add(paddr, 48)
				i -= 48
				if i < 48 {
					break
				}
			}
			seed ^= see1 ^ see2
		}
		for i > 16 {
			seed = _wymix(runtimex.ReadUnaligned64(paddr)^s1, runtimex.ReadUnaligned64(add(paddr, 8))^seed)
			i -= 16
			paddr = add(paddr, 16)
		}
		a = runtimex.ReadUnaligned64(add(paddr, uintptr(i-16)))
		b = runtimex.ReadUnaligned64(add(paddr, uintptr(i-8)))
	}
	a ^= s1
	b ^= seed
	hi, lo := bits.Mul64(a, b)
	a, b = lo, hi
	return _wymix(a^s0^length, b^s1)
}
