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

package wyhash

import (
	"math/bits"
	"reflect"
	"unsafe"

	"github.com/bytedance/gopkg/internal/runtimex"
)

type Digest struct {
	initseed uint64
	seed     uint64
	see1     uint64
	see2     uint64
	tail     [16]byte // tail of the last consumed 48-byte block
	data     [48]byte // buffered, not yet consumed data
	length   int
	total    int
	started  bool // whether seed has been mixed and at least one block consumed
}

// New creates a new Digest that computes the 64-bit wyhash algorithm.
func New(seed uint64) *Digest {
	return &Digest{
		seed:     seed,
		initseed: seed,
		see1:     seed,
		see2:     seed,
	}
}

func NewDefault() *Digest {
	return New(DefaultSeed)
}

// Size always returns 8 bytes.
func (d *Digest) Size() int { return 8 }

// BlockSize always returns 48 bytes.
func (d *Digest) BlockSize() int { return 48 }

// Reset the digest, and the seed will be reset to initseed.
// The initseed is the seed when digest has been created.
func (d *Digest) Reset() {
	d.seed = d.initseed
	d.see1 = d.initseed
	d.see2 = d.initseed
	d.total = 0
	d.length = 0
	d.started = false
}

func (d *Digest) SetSeed(seed uint64)     { d.seed = seed }
func (d *Digest) Seed() uint64            { return d.seed }
func (d *Digest) SetInitSeed(seed uint64) { d.initseed = seed }
func (d *Digest) InitSeed() uint64        { return d.initseed }

// Write (via the embedded io.Writer interface) adds more data to the running hash.
// It never returns an error.
func (d *Digest) Write(input []byte) (int, error) {
	ilen := len(input)
	d.total += ilen

	if d.length+ilen < 48 {
		copy(d.data[d.length:d.length+ilen], input)
		d.length += ilen
		return ilen, nil
	}

	// The seed is mixed exactly once, before the first 48-byte block is
	// consumed (Sum64 delegates to Sum64StringWithSeed when no block was
	// consumed, which performs the same mix itself). As in the reference
	// implementation, see1/see2 start from the mixed seed.
	if !d.started {
		d.seed ^= _wymix(d.seed^s0, s1)
		d.see1 = d.seed
		d.see2 = d.seed
		d.started = true
	}

	// Consume the first 48 bytes if possible.
	if d.length != 0 {
		inputpre := 48 - d.length // the data from input to d.data
		copy(d.data[d.length:], input[:inputpre])
		input = input[inputpre:] // free preceding data

		paddr := unsafe.Pointer(&d.data)
		d.seed = _wymix(runtimex.ReadUnaligned64(paddr)^s1, runtimex.ReadUnaligned64(add(paddr, 8))^d.seed)
		d.see1 = _wymix(runtimex.ReadUnaligned64(add(paddr, 16))^s2, runtimex.ReadUnaligned64(add(paddr, 24))^d.see1)
		d.see2 = _wymix(runtimex.ReadUnaligned64(add(paddr, 32))^s3, runtimex.ReadUnaligned64(add(paddr, 40))^d.see2)
		copy(d.tail[:], d.data[32:48])
		d.length = 0 // free d.data, since it has been consumed
	}

	// If remain input still greater or equal than 48.
	seed, see1, see2 := d.seed, d.see1, d.see2
	for len(input) >= 48 {
		paddr := *(*unsafe.Pointer)(unsafe.Pointer(&input))
		seed = _wymix(runtimex.ReadUnaligned64(paddr)^s1, runtimex.ReadUnaligned64(add(paddr, 8))^seed)
		see1 = _wymix(runtimex.ReadUnaligned64(add(paddr, 16))^s2, runtimex.ReadUnaligned64(add(paddr, 24))^see1)
		see2 = _wymix(runtimex.ReadUnaligned64(add(paddr, 32))^s3, runtimex.ReadUnaligned64(add(paddr, 40))^see2)
		copy(d.tail[:], input[32:48])
		input = input[48:]
	}
	d.seed, d.see1, d.see2 = seed, see1, see2

	// Store remain data(< 48 bytes), d.length == 0 for now.
	if len(input) > 0 {
		copy(d.data[:len(input)], input)
		d.length = len(input)
	}

	return ilen, nil
}

// Sum64 returns the current hash.
func (d *Digest) Sum64() uint64 {
	if !d.started {
		return Sum64StringWithSeed(string(byteToSlice(d.data, d.length)), d.initseed)
	}

	seed := d.seed ^ d.see1 ^ d.see2
	total := uint64(d.total)
	i := uint64(d.length)
	paddr := unsafe.Pointer(&d.data)
	for i > 16 {
		seed = _wymix(runtimex.ReadUnaligned64(paddr)^s1, runtimex.ReadUnaligned64(add(paddr, 8))^seed)
		i -= 16
		paddr = add(paddr, 16)
	}
	// i <= 16. The final reads cover the trailing window
	// [paddr+i-16, paddr+i): when the remaining block is shorter than 16
	// bytes and sits at the start of d.data, its prefix belongs to the tail of
	// the last consumed 48-byte block (as in the reference implementation,
	// which reads p+i-16 unconditionally). Assemble the window so the reads
	// stay in bounds.
	var win [16]byte
	if i < 16 && uintptr(paddr) == uintptr(unsafe.Pointer(&d.data)) {
		copy(win[:16-i], d.tail[i:16])
		copy(win[16-i:], d.data[:i])
		paddr = unsafe.Pointer(&win)
	} else {
		paddr = add(paddr, uintptr(int64(i)-16))
	}
	a := runtimex.ReadUnaligned64(paddr)
	b := runtimex.ReadUnaligned64(add(paddr, 8))
	a ^= s1
	b ^= seed
	hi, lo := bits.Mul64(a, b)
	a, b = lo, hi
	return _wymix(a^s0^total, b^s1)
}

// Sum appends the current hash to b and returns the resulting slice.
func (d *Digest) Sum(b []byte) []byte {
	s := d.Sum64()
	return append(
		b,
		byte(s>>56),
		byte(s>>48),
		byte(s>>40),
		byte(s>>32),
		byte(s>>24),
		byte(s>>16),
		byte(s>>8),
		byte(s),
	)
}

func byteToSlice(b [48]byte, length int) []byte {
	var res []byte
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&res))
	bh.Data = uintptr(unsafe.Pointer(&b))
	bh.Len = length
	bh.Cap = length
	return res
}
