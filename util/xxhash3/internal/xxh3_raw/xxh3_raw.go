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

package xxh3_raw

import (
	"math/bits"
	"unsafe"
)

const (
	_stripe = 64
	_block  = 1024

	prime32_1 = 2654435761
	prime32_2 = 2246822519
	prime32_3 = 3266489917

	prime64_1 = 11400714785074694791
	prime64_2 = 14029467366897019727
	prime64_3 = 1609587929392839161
	prime64_4 = 9650029242287828579
	prime64_5 = 2870177450012600261
)

var xsecret = unsafe.Pointer(&[192]uint8{
	/* 	0 	*/ 0xb8, 0xfe, 0x6c, 0x39, 0x23, 0xa4, 0x4b, 0xbe, 0x7c, 0x01, 0x81, 0x2c, 0xf7, 0x21, 0xad, 0x1c,
	/* 	16 	*/ 0xde, 0xd4, 0x6d, 0xe9, 0x83, 0x90, 0x97, 0xdb, 0x72, 0x40, 0xa4, 0xa4, 0xb7, 0xb3, 0x67, 0x1f,
	/* 	32 	*/ 0xcb, 0x79, 0xe6, 0x4e, 0xcc, 0xc0, 0xe5, 0x78, 0x82, 0x5a, 0xd0, 0x7d, 0xcc, 0xff, 0x72, 0x21,
	/* 	48 	*/ 0xb8, 0x08, 0x46, 0x74, 0xf7, 0x43, 0x24, 0x8e, 0xe0, 0x35, 0x90, 0xe6, 0x81, 0x3a, 0x26, 0x4c,
	/* 	64 	*/ 0x3c, 0x28, 0x52, 0xbb, 0x91, 0xc3, 0x00, 0xcb, 0x88, 0xd0, 0x65, 0x8b, 0x1b, 0x53, 0x2e, 0xa3,
	/* 	80 	*/ 0x71, 0x64, 0x48, 0x97, 0xa2, 0x0d, 0xf9, 0x4e, 0x38, 0x19, 0xef, 0x46, 0xa9, 0xde, 0xac, 0xd8,
	/* 	96 	*/ 0xa8, 0xfa, 0x76, 0x3f, 0xe3, 0x9c, 0x34, 0x3f, 0xf9, 0xdc, 0xbb, 0xc7, 0xc7, 0x0b, 0x4f, 0x1d,
	/* 	112	*/ 0x8a, 0x51, 0xe0, 0x4b, 0xcd, 0xb4, 0x59, 0x31, 0xc8, 0x9f, 0x7e, 0xc9, 0xd9, 0x78, 0x73, 0x64,
	/* 	128	*/ 0xea, 0xc5, 0xac, 0x83, 0x34, 0xd3, 0xeb, 0xc3, 0xc5, 0x81, 0xa0, 0xff, 0xfa, 0x13, 0x63, 0xeb,
	/* 	144	*/ 0x17, 0x0d, 0xdd, 0x51, 0xb7, 0xf0, 0xda, 0x49, 0xd3, 0x16, 0x55, 0x26, 0x29, 0xd4, 0x68, 0x9e,
	/* 	160	*/ 0x2b, 0x16, 0xbe, 0x58, 0x7d, 0x47, 0xa1, 0xfc, 0x8f, 0xf8, 0xb8, 0xd1, 0x7a, 0xd0, 0x31, 0xce,
	/* 	176	*/ 0x45, 0xcb, 0x3a, 0x8f, 0x95, 0x16, 0x04, 0x28, 0xaf, 0xd7, 0xfb, 0xca, 0xbb, 0x4b, 0x40, 0x7e,
})

// Hash returns the hash value of the byte slice in 64bits.
func Hash(data []byte) uint64 {
	length := uint64(len(data))
	xinput := *(*unsafe.Pointer)(unsafe.Pointer(&data))

	if length > 240 {
		return hashLarge(xinput, length)
	} else if length > 128 {
		return xxh3Len129To240_64b(xinput, length)
	} else if length > 16 {
		return xxh3Len17To128_64b(xinput, length)
	} else {
		return xxh3Len0To16_64b(xinput, length)
	}
}

// HashString returns the hash value of the string in 64bits.
func HashString(s string) uint64 {
	return Hash([]byte(s))
}

// Hash128 returns the hash value of the byte slice in 128bits.
func Hash128(data []byte) [2]uint64 {
	length := uint64(len(data))
	xinput := *(*unsafe.Pointer)(unsafe.Pointer(&data))

	if length > 240 {
		return hashLarge128(xinput, length)
	} else if length > 128 {
		return xxh3Len129To240_128b(xinput, length)
	} else if length > 16 {
		return xxh3Len17To128_128b(xinput, length)
	} else {
		return xxh3Len0To16_128b(xinput, length)
	}
}

// Hash128String returns the hash value of the string in 128bits.
func Hash128String(s string) [2]uint64 {
	return Hash128([]byte(s))
}

func xxh3Len0To16_64b(xinput unsafe.Pointer, length uint64) uint64 {
	if length > 8 {
		inputlo := Read8(xinput, 0) ^ (Read8(xsecret, 24)) ^ Read8(xsecret, 32)
		inputhi := Read8(xinput, uintptr(length-8)) ^ (Read8(xsecret, 40)) ^ Read8(xsecret, 48)
		acc := length + bits.ReverseBytes64(inputlo) + inputhi + mix(inputlo, inputhi)
		return xxh3Avalanche(acc)
	} else if length >= 4 {
		input1 := Read4(xinput, 0)
		input2 := Read4(xinput, uintptr(length-4))
		input64 := input2 + input1<<32
		keyed := input64 ^ (Read8(xsecret, 8)) ^ Read8(xsecret, 16)
		return xxh3RRMXMX(keyed, length)
	} else if length > 0 {
		q := (*[4]byte)(xinput)
		combined := (uint64(q[0]) << 16) | (uint64(q[length>>1]) << 24) | (uint64(q[length-1]) << 0) | length<<8
		combined ^= Read4(xsecret, 0) ^ Read4(xsecret, 4)
		return xxh64Avalanche(combined)
	} else {
		return xxh64Avalanche(Read8(xsecret, 56) ^ Read8(xsecret, 64))
	}
}

func xxh3Len17To128_64b(xinput unsafe.Pointer, length uint64) uint64 {
	acc := length * prime64_1
	if length > 32 {
		if length > 64 {
			if length > 96 {
				acc += mix(Read8(xinput, 48)^Read8(xsecret, 96), Read8(xinput, 56)^Read8(xsecret, 104))
				acc += mix(Read8(xinput, uintptr(length-64))^Read8(xsecret, 112), Read8(xinput, uintptr(length-56))^Read8(xsecret, 120))
			}
			acc += mix(Read8(xinput, 32)^Read8(xsecret, 64), Read8(xinput, 40)^Read8(xsecret, 72))
			acc += mix(Read8(xinput, uintptr(length-48))^Read8(xsecret, 80), Read8(xinput, uintptr(length-40))^Read8(xsecret, 88))
		}
		acc += mix(Read8(xinput, 16)^Read8(xsecret, 32), Read8(xinput, 24)^Read8(xsecret, 40))
		acc += mix(Read8(xinput, uintptr(length-32))^Read8(xsecret, 48), Read8(xinput, uintptr(length-24))^Read8(xsecret, 56))
	}
	acc += mix(Read8(xinput, 0)^Read8(xsecret, 0), Read8(xinput, 8)^Read8(xsecret, 8))
	acc += mix(Read8(xinput, uintptr(length-16))^Read8(xsecret, 16), Read8(xinput, uintptr(length-8))^Read8(xsecret, 24))

	return xxh3Avalanche(acc)
}

func xxh3Len129To240_64b(xinput unsafe.Pointer, length uint64) uint64 {

	acc := length * prime64_1

	acc += mix(Read8(xinput, 16*0)^Read8(xsecret, 16*0), Read8(xinput, 16*0+8)^Read8(xsecret, 16*0+8))
	acc += mix(Read8(xinput, 16*1)^Read8(xsecret, 16*1), Read8(xinput, 16*1+8)^Read8(xsecret, 16*1+8))
	acc += mix(Read8(xinput, 16*2)^Read8(xsecret, 16*2), Read8(xinput, 16*2+8)^Read8(xsecret, 16*2+8))
	acc += mix(Read8(xinput, 16*3)^Read8(xsecret, 16*3), Read8(xinput, 16*3+8)^Read8(xsecret, 16*3+8))
	acc += mix(Read8(xinput, 16*4)^Read8(xsecret, 16*4), Read8(xinput, 16*4+8)^Read8(xsecret, 16*4+8))
	acc += mix(Read8(xinput, 16*5)^Read8(xsecret, 16*5), Read8(xinput, 16*5+8)^Read8(xsecret, 16*5+8))
	acc += mix(Read8(xinput, 16*6)^Read8(xsecret, 16*6), Read8(xinput, 16*6+8)^Read8(xsecret, 16*6+8))
	acc += mix(Read8(xinput, 16*7)^Read8(xsecret, 16*7), Read8(xinput, 16*7+8)^Read8(xsecret, 16*7+8))

	acc = xxh3Avalanche(acc)
	nbRounds := length >> 4

	for i := uint64(8); i < nbRounds; i++ {
		acc += mix(Read8(xinput, uintptr(16*i))^Read8(xsecret, uintptr(16*i-125)), Read8(xinput, uintptr(16*i+8))^Read8(xsecret, uintptr(16*i-117)))
	}

	acc += mix(Read8(xinput, uintptr(length-16))^Read8(xsecret, 119), Read8(xinput, uintptr(length-8))^Read8(xsecret, uintptr(127)))

	return xxh3Avalanche(acc)
}

func hashLarge(p unsafe.Pointer, length uint64) (acc uint64) {
	acc = length * prime64_1

	xacc := [8]uint64{
		prime32_3, prime64_1, prime64_2, prime64_3,
		prime64_4, prime32_2, prime64_5, prime32_1}

	accumScalar(&xacc, p, xsecret, length)
	//merge xacc
	acc += mix(xacc[0]^Read8(xsecret, 11), xacc[1]^Read8(xsecret, 19))
	acc += mix(xacc[2]^Read8(xsecret, 27), xacc[3]^Read8(xsecret, 35))
	acc += mix(xacc[4]^Read8(xsecret, 43), xacc[5]^Read8(xsecret, 51))
	acc += mix(xacc[6]^Read8(xsecret, 59), xacc[7]^Read8(xsecret, 67))

	return xxh3Avalanche(acc)
}

func xxh3Len0To16_128b(xinput unsafe.Pointer, length uint64) [2]uint64 {

	if length > 8 {
		bitflipl := Read8(xsecret, 32) ^ Read8(xsecret, 40)
		bitfliph := Read8(xsecret, 48) ^ Read8(xsecret, 56)
		inputLow := Read8(xinput, 0)
		inputHigh := Read8(xinput, uintptr(length)-8)
		m128High64, m128Low64 := bits.Mul64(inputLow^inputHigh^bitflipl, prime64_1)

		m128Low64 += uint64(length-1) << 54
		inputHigh ^= bitfliph

		m128High64 += inputHigh + uint64(uint32(inputHigh))*(prime32_2-1)
		m128Low64 ^= bits.ReverseBytes64(m128High64)

		h128High64, h128Low64 := bits.Mul64(m128Low64, prime64_2)
		h128High64 += m128High64 * prime64_2

		h128Low64 = xxh3Avalanche(h128Low64)
		h128High64 = xxh3Avalanche(h128High64)

		return [2]uint64{h128High64, h128Low64}

	} else if length >= 4 {
		inputLow := Read4(xinput, 0)
		inputHigh := Read4(xinput, uintptr(length)-4)
		input64 := inputLow + (uint64(inputHigh) << 32)
		bitflip := Read8(xsecret, 16) ^ Read8(xsecret, 24)
		keyed := input64 ^ bitflip

		m128High64, m128Low64 := bits.Mul64(keyed, prime64_1+(length)<<2)
		m128High64 += m128Low64 << 1
		m128Low64 ^= m128High64 >> 3

		m128Low64 ^= m128Low64 >> 35
		m128Low64 *= 0x9fb21c651e98df25
		m128Low64 ^= m128Low64 >> 28

		m128High64 = xxh3Avalanche(m128High64)
		return [2]uint64{m128High64, m128Low64}

	} else if length >= 1 {
		q := (*[4]byte)(xinput)
		combinedl := (uint64(q[0]) << 16) | (uint64(q[length>>1]) << 24) | (uint64(q[length-1]) << 0) | length<<8
		combinedh := uint64(bits.RotateLeft32(bits.ReverseBytes32(uint32(combinedl)), 13))

		bitflipl := Read4(xsecret, 0) ^ Read4(xsecret, 4)
		bitfliph := Read4(xsecret, 8) ^ Read4(xsecret, 12)

		keyedLow := combinedl ^ bitflipl
		keyedHigh := combinedh ^ bitfliph

		keyedLow = combinedl ^ bitflipl
		keyedHigh = combinedh ^ bitfliph

		h128Low64 := xxh64Avalanche(keyedLow)
		h128High64 := xxh64Avalanche(keyedHigh)
		return [2]uint64{h128High64, h128Low64}
	}
	bitflipl := Read8(xsecret, 64) ^ Read8(xsecret, 72)
	bitfliph := Read8(xsecret, 80) ^ Read8(xsecret, 88)

	h128High64 := xxh64Avalanche(bitfliph)
	h128Low64 := xxh64Avalanche(bitflipl)

	return [2]uint64{h128High64, h128Low64}
}

func xxh3Len17To128_128b(xinput unsafe.Pointer, length uint64) [2]uint64 {

	accHigh := uint64(0)
	accLow := length * prime64_1

	if length > 32 {
		if length > 64 {
			if length > 96 {
				accLow += mix(Read8(xinput, 48)^Read8(xsecret, 96), Read8(xinput, 56)^Read8(xsecret, 104))
				accLow ^= Read8(xinput, uintptr(length-64)) + Read8(xinput, uintptr(length-56))
				accHigh += mix(Read8(xinput, uintptr(length-64))^Read8(xsecret, 112), Read8(xinput, uintptr(length-56))^Read8(xsecret, 120))
				accHigh ^= Read8(xinput, 48) + Read8(xinput, 56)
			}
			accLow += mix(Read8(xinput, 32)^Read8(xsecret, 64), Read8(xinput, 40)^Read8(xsecret, 72))
			accLow ^= Read8(xinput, uintptr(length-48)) + Read8(xinput, uintptr(length-40))
			accHigh += mix(Read8(xinput, uintptr(length-48))^Read8(xsecret, 80), Read8(xinput, uintptr(length-40))^Read8(xsecret, 88))
			accHigh ^= Read8(xinput, 32) + Read8(xinput, 40)
		}
		accLow += mix(Read8(xinput, 16)^Read8(xsecret, 32), Read8(xinput, 3*8)^Read8(xsecret, 40))
		accLow ^= Read8(xinput, uintptr(length-32)) + Read8(xinput, uintptr(length-3*8))
		accHigh += mix(Read8(xinput, uintptr(length-32))^Read8(xsecret, 48), Read8(xinput, uintptr(length-3*8))^Read8(xsecret, 56))
		accHigh ^= Read8(xinput, 16) + Read8(xinput, 3*8)
	}

	accLow += mix(Read8(xinput, 0)^Read8(xsecret, 0), Read8(xinput, 8)^Read8(xsecret, 8))
	accLow ^= Read8(xinput, uintptr(length-16)) + Read8(xinput, uintptr(length-8))
	accHigh += mix(Read8(xinput, uintptr(length-16))^Read8(xsecret, 16), Read8(xinput, uintptr(length-8))^Read8(xsecret, 24))
	accHigh ^= Read8(xinput, 0) + Read8(xinput, 8)

	h128Low := accHigh + accLow
	h128High := (accLow * prime64_1) + (accHigh * prime64_4) + (length * prime64_2)

	h128Low = xxh3Avalanche(h128Low)
	h128High = -xxh3Avalanche(h128High)

	return [2]uint64{h128High, h128Low}
}

func xxh3Len129To240_128b(xinput unsafe.Pointer, length uint64) [2]uint64 {
	nbRounds := length &^ 31 / 32
	accLow64 := length * prime64_1
	accHigh64 := uint64(0)

	for i := 0; i < 4; i++ {
		accLow64 += mix(Read8(xinput, uintptr(32*i))^Read8(xsecret, uintptr(32*i)), Read8(xinput, uintptr(32*i+8))^Read8(xsecret, uintptr(32*i+8)))
		accLow64 ^= Read8(xinput, uintptr(32*i+16)) + Read8(xinput, uintptr(32*i+24))
		accHigh64 += mix(Read8(xinput, uintptr(32*i+16))^Read8(xsecret, uintptr(32*i+16)), Read8(xinput, uintptr(32*i)+24)^Read8(xsecret, uintptr(32*i+24)))
		accHigh64 ^= Read8(xinput, uintptr(32*i)) + Read8(xinput, uintptr(32*i)+8)
	}

	accLow64 = xxh3Avalanche(accLow64)
	accHigh64 = xxh3Avalanche(accHigh64)

	for i := uint64(4); i < nbRounds; i++ {
		accHigh64 += mix(Read8(xinput, uintptr(32*i+16))^Read8(xsecret, uintptr(32*i-109)), Read8(xinput, uintptr(32*i)+24)^Read8(xsecret, uintptr(32*i-101)))
		accHigh64 ^= Read8(xinput, uintptr(32*i)) + Read8(xinput, uintptr(32*i)+8)

		accLow64 += mix(Read8(xinput, uintptr(32*i))^Read8(xsecret, uintptr(32*i-125)), Read8(xinput, uintptr(32*i+8))^Read8(xsecret, uintptr(32*i-117)))
		accLow64 ^= Read8(xinput, uintptr(32*i+16)) + Read8(xinput, uintptr(32*i+24))
	}

	// last 32 bytes
	accLow64 += mix(Read8(xinput, uintptr(length-16))^Read8(xsecret, 103), Read8(xinput, uintptr(length-8))^Read8(xsecret, 111))
	accLow64 ^= Read8(xinput, uintptr(length-32)) + Read8(xinput, uintptr(length-24))
	accHigh64 += mix(Read8(xinput, uintptr(length-32))^Read8(xsecret, 119), Read8(xinput, uintptr(length-24))^Read8(xsecret, 127))
	accHigh64 ^= Read8(xinput, uintptr(length-16)) + Read8(xinput, uintptr(length-8))

	accHigh64, accLow64 = (accLow64*prime64_1)+(accHigh64*prime64_4)+(length*prime64_2), accHigh64+accLow64

	accLow64 = xxh3Avalanche(accLow64)
	accHigh64 = -xxh3Avalanche(accHigh64)

	return [2]uint64{accHigh64, accLow64}
}

func hashLarge128(p unsafe.Pointer, length uint64) (acc [2]uint64) {
	acc[1] = length * prime64_1
	acc[0] = ^(length * prime64_2)

	xacc := [8]uint64{
		prime32_3, prime64_1, prime64_2, prime64_3,
		prime64_4, prime32_2, prime64_5, prime32_1}

	accumScalar(&xacc, p, xsecret, length)
	// merge xacc
	acc[1] += mix(xacc[0]^Read8(xsecret, 11), xacc[1]^Read8(xsecret, 19))
	acc[1] += mix(xacc[2]^Read8(xsecret, 27), xacc[3]^Read8(xsecret, 35))
	acc[1] += mix(xacc[4]^Read8(xsecret, 43), xacc[5]^Read8(xsecret, 51))
	acc[1] += mix(xacc[6]^Read8(xsecret, 59), xacc[7]^Read8(xsecret, 67))

	acc[1] = xxh3Avalanche(acc[1])

	acc[0] += mix(xacc[0]^Read8(xsecret, 117), xacc[1]^Read8(xsecret, 125))
	acc[0] += mix(xacc[2]^Read8(xsecret, 133), xacc[3]^Read8(xsecret, 141))
	acc[0] += mix(xacc[4]^Read8(xsecret, 149), xacc[5]^Read8(xsecret, 157))
	acc[0] += mix(xacc[6]^Read8(xsecret, 165), xacc[7]^Read8(xsecret, 173))
	acc[0] = xxh3Avalanche(acc[0])

	return acc
}

func accumScalar(xacc *[8]uint64, xinput, xsecret unsafe.Pointer, l uint64) {
	j := uint64(0)

	// Loops over block, process 16*8*8=1024 bytes of data each iteration
	for ; j < (l-1)/1024; j++ {
		k := xsecret
		for i := 0; i < 16; i++ {
			for j := uintptr(0); j < 8; j++ {
				dataVec := Read8(xinput, 8*j)
				keyVec := dataVec ^ Read8(k, 8*j)
				xacc[j^1] += dataVec
				xacc[j] += (keyVec & 0xffffffff) * (keyVec >> 32)
			}
			xinput, k = unsafe.Pointer(uintptr(xinput)+_stripe), unsafe.Pointer(uintptr(k)+8)
		}

		// scramble xacc
		for j := uintptr(0); j < 8; j++ {
			xacc[j] ^= xacc[j] >> 47
			xacc[j] ^= Read8(xsecret, 128+8*j)
			xacc[j] *= prime32_1
		}
	}
	l -= _block * j

	// last partial block (1024 bytes)
	if l > 0 {
		k := xsecret
		i := uint64(0)
		for ; i < (l-1)/_stripe; i++ {
			for j := uintptr(0); j < 8; j++ {
				dataVec := Read8(xinput, 8*j)
				keyVec := dataVec ^ Read8(k, 8*j)
				xacc[j^1] += dataVec
				xacc[j] += (keyVec & 0xffffffff) * (keyVec >> 32)
			}
			xinput, k = unsafe.Pointer(uintptr(xinput)+_stripe), unsafe.Pointer(uintptr(k)+8)
		}
		l -= _stripe * i

		// last stripe (64 bytes)
		if l > 0 {
			xinput = unsafe.Pointer(uintptr(xinput) - uintptr(_stripe-l))
			k = unsafe.Pointer(uintptr(xsecret) + 121)

			for j := uintptr(0); j < 8; j++ {
				dataVec := Read8(xinput, 8*j)
				keyVec := dataVec ^ Read8(k, 8*j)
				xacc[j^1] += dataVec
				xacc[j] += (keyVec & 0xffffffff) * (keyVec >> 32)
			}
		}
	}
}

func mix(a, b uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	return hi ^ lo
}
func xxh3RRMXMX(h64 uint64, length uint64) uint64 {
	h64 ^= bits.RotateLeft64(h64, 49) ^ bits.RotateLeft64(h64, 24)
	h64 *= 0x9fb21c651e98df25
	h64 ^= (h64 >> 35) + length
	h64 *= 0x9fb21c651e98df25
	h64 ^= (h64 >> 28)
	return h64
}

func xxh64Avalanche(h64 uint64) uint64 {
	h64 ^= h64 >> 33
	h64 *= prime64_2
	h64 ^= h64 >> 29
	h64 *= prime64_3
	h64 ^= h64 >> 32
	return h64
}

func xxh3Avalanche(x uint64) uint64 {
	x ^= x >> 37
	x *= 0x165667919e3779f9
	x ^= x >> 32
	return x
}

func Read8(p unsafe.Pointer, offset uintptr) uint64 {
	p = unsafe.Pointer(uintptr(p) + offset)
	q := (*[8]byte)(p)
	return uint64(q[0]) | uint64(q[1])<<8 | uint64(q[2])<<16 | uint64(q[3])<<24 | uint64(q[4])<<32 | uint64(q[5])<<40 | uint64(q[6])<<48 | uint64(q[7])<<56
}

func Read4(p unsafe.Pointer, offset uintptr) uint64 {
	p = unsafe.Pointer(uintptr(p) + offset)
	q := (*[4]byte)(p)
	return uint64(q[0]) | uint64(q[1])<<8 | uint64(q[2])<<16 | uint64(q[3])<<24
}
