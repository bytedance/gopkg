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

//go:build ignore
// +build ignore

package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func AVX2() {

	primeData := GLOBL("prime_avx", RODATA|NOPTR)
	DATA(0, U32(2654435761))
	TEXT("accumAVX2", NOSPLIT, "func(acc *[8]uint64, xinput, xsecret *byte, len uint64)")

	acc := Mem{Base: Load(Param("acc"), GP64())}
	xinput := Mem{Base: Load(Param("xinput"), GP64())}
	xsecret := Mem{Base: Load(Param("xsecret"), GP64())}
	skey := Mem{Base: Load(Param("xsecret"), GP64())}
	plen := Load(Param("len"), GP64())
	prime := YMM()
	a := [...]VecVirtual{YMM(), YMM()}

	VMOVDQU(acc.Offset(0x00), a[0])
	VMOVDQU(acc.Offset(0x20), a[1])
	VPBROADCASTQ(primeData, prime)

	// Loops over block, process 16*8*8=1024 bytes of data each iteration
	Label("accumBlock")
	{
		CMPQ(plen, U32(1024))
		JLE(LabelRef("accumStripe"))

		for i := 0; i < 8; i++ {
			y0, y1, y2, y3, y4, y5, y6, y7, y8 := YMM(), YMM(), YMM(), YMM(), YMM(), YMM(), YMM(), YMM(), YMM()
			//data_vec    = xinput[i]
			VMOVDQU(xinput.Offset(128*i), y0)
			//key_vec     = xsecret[i]
			VMOVDQU(xsecret.Offset(16*i), y1)
			VMOVDQU(xinput.Offset(128*i+0x20), y3)
			VMOVDQU(xsecret.Offset(16*i+0x20), y4)
			VMOVDQU(xinput.Offset(128*i+0x40), y5)
			VMOVDQU(xsecret.Offset(16*i+0x8), y6)
			VMOVDQU(xinput.Offset(128*i+0x60), y7)
			VMOVDQU(xsecret.Offset(16*i+0x28), y8)

			// data_key    = data_vec ^ key_vec
			VPXOR(y0, y1, y1)
			// data_key_lo = data_key >> 32
			VPSRLQ(Imm(0x20), y1, y2)
			VPSHUFD(Imm(0x4e), y0, y0)
			// product     = (data_key & 0xffffffff) * (data_key_lo & 0xffffffff)
			VPMULUDQ(y1, y2, y2)
			// xacc[i] += swap(data_vec)
			VPADDQ(a[0], y0, a[0])
			// xacc[i] += product
			VPADDQ(a[0], y2, a[0])

			VPXOR(y3, y4, y4)
			VPSRLQ(Imm(0x20), y4, y2)
			VPSHUFD(Imm(0x4e), y3, y3)
			VPMULUDQ(y4, y2, y2)
			VPADDQ(a[1], y3, a[1])
			VPADDQ(a[1], y2, a[1])

			VPXOR(y5, y6, y6)
			VPSRLQ(Imm(0x20), y6, y2)
			VPSHUFD(Imm(0x4e), y5, y5)
			VPMULUDQ(y6, y2, y2)
			VPADDQ(a[0], y5, a[0])
			VPADDQ(a[0], y2, a[0])

			VPXOR(y7, y8, y8)
			VPSRLQ(Imm(0x20), y8, y2)
			VPSHUFD(Imm(0x4e), y7, y7)
			VPMULUDQ(y8, y2, y2)
			VPADDQ(a[1], y7, a[1])
			VPADDQ(a[1], y2, a[1])
		}

		ADDQ(U32(16*64), xinput.Base)
		SUBQ(U32(16*64), plen)

		y0, y1 := YMM(), YMM()

		// xacc[i] ^= (xacc[i] >> 47)
		VPSRLQ(Imm(0x2f), a[0], y0)
		VPXOR(a[0], y0, y0)
		// xacc[i] ^= xsecret
		VPXOR(xsecret.Offset(0x80), y0, y0)
		VPMULUDQ(prime, y0, y1)
		// xacc[i] *= prime;
		VPSRLQ(Imm(0x20), y0, y0)
		VPMULUDQ(prime, y0, y0)
		VPSLLQ(Imm(0x20), y0, y0)
		VPADDQ(y1, y0, a[0])

		VPSRLQ(Imm(0x2f), a[1], y0)
		VPXOR(a[1], y0, y0)
		VPXOR(xsecret.Offset(0xa0), y0, y0)
		VPMULUDQ(prime, y0, y1)
		VPSRLQ(Imm(0x20), y0, y0)
		VPMULUDQ(prime, y0, y0)
		VPSLLQ(Imm(0x20), y0, y0)
		VPADDQ(y1, y0, a[1])
		JMP(LabelRef("accumBlock"))
	}

	// last partial block (64 bytes)
	Label("accumStripe")
	{
		CMPQ(plen, Imm(64))
		JLE(LabelRef("accumLastStripe"))

		y0, y1, y2, y3, y4 := YMM(), YMM(), YMM(), YMM(), YMM()
		VMOVDQU(xinput.Offset(0), y0)
		VMOVDQU(skey.Offset(0), y1)
		VMOVDQU(xinput.Offset(0x20), y3)
		VMOVDQU(skey.Offset(0x20), y4)

		// data_key    = data_vec ^ key_vec
		VPXOR(y0, y1, y1)
		// data_key_lo = data_key >> 32
		VPSRLQ(Imm(0x20), y1, y2)
		VPSHUFD(Imm(0x4e), y0, y0)
		// product     = (data_key & 0xffffffff) * (data_key_lo & 0xffffffff)
		VPMULUDQ(y1, y2, y2)
		// xacc[i] += swap(data_vec)
		VPADDQ(a[0], y0, a[0])
		// xacc[i] += product
		VPADDQ(a[0], y2, a[0])

		VPXOR(y3, y4, y4)
		VPSRLQ(Imm(0x20), y4, y2)
		VPMULUDQ(y4, y2, y2)
		VPSHUFD(Imm(0x4e), y3, y3)
		VPADDQ(a[1], y3, a[1])
		VPADDQ(a[1], y2, a[1])

		ADDQ(U32(64), xinput.Base)
		SUBQ(U32(64), plen)
		ADDQ(U32(8), skey.Base)

		JMP(LabelRef("accumStripe"))
	}

	// last stripe 16 bytes
	Label("accumLastStripe")
	{
		CMPQ(plen, Imm(0))
		JE(LabelRef("return"))

		SUBQ(Imm(64), xinput.Base)
		ADDQ(plen, xinput.Base)

		y0, y1, y2, y3, y4 := YMM(), YMM(), YMM(), YMM(), YMM()
		VMOVDQU(xinput.Offset(0), y0)
		VMOVDQU(xsecret.Offset(0x79), y1)
		VMOVDQU(xinput.Offset(0x20), y3)
		VMOVDQU(xsecret.Offset(0x79+0x20), y4)

		// data_key    = data_vec ^ key_vec
		VPXOR(y0, y1, y1)
		// data_key_lo = data_key >> 32
		VPSRLQ(Imm(0x20), y1, y2)
		VPSHUFD(Imm(0x4e), y0, y0)
		// product     = (data_key & 0xffffffff) * (data_key_lo & 0xffffffff)
		VPMULUDQ(y1, y2, y2)
		// xacc[i] += swap(data_vec)
		VPADDQ(a[0], y0, a[0])
		// xacc[i] += product
		VPADDQ(a[0], y2, a[0])

		VPXOR(y3, y4, y4)
		VPSRLQ(Imm(0x20), y4, y2)
		VPMULUDQ(y4, y2, y2)
		VPSHUFD(Imm(0x4e), y3, y3)
		VPADDQ(a[1], y3, a[1])
		VPADDQ(a[1], y2, a[1])
	}

	Label("return")
	{
		VMOVDQU(a[0], acc.Offset(0x00))
		VMOVDQU(a[1], acc.Offset(0x20))
		RET()
	}

	Generate()
}
