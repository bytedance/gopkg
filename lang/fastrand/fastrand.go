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

// Package fastrand is the fastest pseudorandom number generator in Go(multiple-cores).
package fastrand

import (
	"unsafe"

	"github.com/bytedance/gopkg/internal/runtimex"
)

// Uint32 returns a pseudo-random 32-bit value as a uint32.
func Uint32() uint32 {
	return runtimex.Fastrand()
}

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func Uint64() uint64 {
	return (uint64(runtimex.Fastrand()) << 32) | uint64(runtimex.Fastrand())
}

// Int returns a non-negative pseudo-random int.
func Int() int {
	// EQ
	u := uint(Int63())
	return int(u << 1 >> 1) // clear sign bit if int == int32
}

// Int31 returns a non-negative pseudo-random 31-bit integer as an int32.
func Int31() int32 { return int32(Uint32() & (1<<31 - 1)) }

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func Int63() int64 {
	// EQ
	return int64(Uint64() & (1<<63 - 1))
}

// Int63n returns, as an int64, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func Int63n(n int64) int64 {
	// EQ
	if n <= 0 {
		panic("invalid argument to Int63n")
	}
	if n&(n-1) == 0 { // n is power of two, can mask
		return Int63() & (n - 1)
	}
	max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := Int63()
	for v > max {
		v = Int63()
	}
	return v % n
}

// Int31n returns, as an int32, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
// For implementation details, see:
// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction
func Int31n(n int32) int32 {
	// EQ
	if n <= 0 {
		panic("invalid argument to Int31n")
	}
	v := Uint32()
	prod := uint64(v) * uint64(n)
	low := uint32(prod)
	if low < uint32(n) {
		thresh := uint32(-n) % uint32(n)
		for low < thresh {
			v = Uint32()
			prod = uint64(v) * uint64(n)
			low = uint32(prod)
		}
	}
	return int32(prod >> 32)
}

// Intn returns, as an int, a non-negative pseudo-random number in [0,n).
// It panics if n <= 0.
func Intn(n int) int {
	// EQ
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	if n <= 1<<31-1 {
		return int(Int31n(int32(n)))
	}
	return int(Int63n(int64(n)))
}

func Float64() float64 {
	// EQ
	return float64(Int63n(1<<53)) / (1 << 53)
}

func Float32() float32 {
	// EQ
	return float32(Int31n(1<<24)) / (1 << 24)
}

// Uint32n returns a pseudo-random number in [0,n).
//go:nosplit
func Uint32n(n uint32) uint32 {
	// This is similar to Uint32() % n, but faster.
	// See https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
	return uint32(uint64(Uint32()) * uint64(n) >> 32)
}

// Uint64n returns a pseudo-random number in [0,n).
func Uint64n(n uint64) uint64 {
	return Uint64() % n
}

// Read generates len(p) random bytes and writes them into p.
// It always returns len(p) and a nil error.
// It is safe for concurrent use.
func Read(p []byte) (int, error) {
	l := len(p)

	// Used for local XORSHIFT.
	var tmp [2]uint32
	tmp[0], tmp[1] = Uint32(), Uint32()

	if l >= 4 {
		var i int
		uint32p := *(*[]uint32)(unsafe.Pointer(&p))
		s1, s0 := tmp[0], tmp[1]
		for l >= 4 {
			// Local XORSHIFT.
			s1 ^= s1 << 17
			s1 = s1 ^ s0 ^ s1>>7 ^ s0>>16
			s0, s1 = s1, s0

			uint32p[i] = s0 + s1
			i++
			l -= 4
		}
		tmp[0], tmp[1] = s1, s0
	}

	if l > 0 {
		// Local XORSHIFT.
		// We don't need to save s0 and s1(tmp is not used).
		s1, s0 := tmp[0], tmp[1]
		s1 ^= s1 << 17
		s1 = s1 ^ s0 ^ s1>>7 ^ s0>>16

		r := s0 + s1
		for l > 0 {
			p[len(p)-l] = byte(r >> (l * 8))
			l--
		}
	}

	return len(p), nil
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func Shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}
	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(Int31n(int32(i + 1)))
		swap(i, j)
	}
}

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers
// in the half-open interval [0,n).
func Perm(n int) []int {
	m := make([]int, n)
	for i := 1; i < n; i++ {
		j := Intn(i + 1)
		m[i] = m[j]
		m[j] = i
	}
	return m
}
