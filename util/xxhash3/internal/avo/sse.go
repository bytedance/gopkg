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

func SSE2() {

	primeData := GLOBL("prime_sse", RODATA|NOPTR)
	DATA(0, U32(2654435761))
	DATA(4, U32(2654435761))
	DATA(8, U32(2654435761))
	DATA(12, U32(2654435761))

	TEXT("accumSSE2", NOSPLIT, "func(acc *[8]uint64, xinput, xsecret *byte, len uint64)")

	acc := Mem{Base: Load(Param("acc"), GP64())}
	xinput := Mem{Base: Load(Param("xinput"), GP64())}
	xsecret := Mem{Base: Load(Param("xsecret"), GP64())}
	skey := Mem{Base: Load(Param("xsecret"), GP64())}
	plen := Load(Param("len"), GP64())
	prime := XMM()
	a := [4]VecVirtual{XMM(), XMM(), XMM(), XMM()}

	MOVOU(acc.Offset(0x00), a[0])
	MOVOU(acc.Offset(0x10), a[1])
	MOVOU(acc.Offset(0x20), a[2])
	MOVOU(acc.Offset(0x30), a[3])
	MOVOU(primeData, prime)

	// Loops over block, process 16*8*8=1024 bytes of data each iteration
	Label("accumBlock")
	{
		CMPQ(plen, U32(1024))
		JLE(LabelRef("accumStripe"))

		for i := 0; i < 0x10; i++ {
			for n := 0; n < 4; n++ {
				x0, x1, x2 := XMM(), XMM(), XMM()
				// data_vec    = xinput[i]
				MOVOU(xinput.Offset(0x40*i+n*0x10), x0)
				// key_vec     = xsecret[i]
				MOVOU(xsecret.Offset(8*i+n*0x10), x1)
				// data_key    = data_vec ^ key_vec
				PXOR(x0, x1)
				// product     = (data_key & 0xffffffff) * (data_key_lo & 0xffffffff)
				PSHUFD(Imm(0x31), x1, x2)
				PMULULQ(x1, x2)
				// xacc[i] += swap(data_vec)
				PSHUFD(Imm(0x4e), x0, x0)
				PADDQ(x0, a[n])
				// xacc[i] += product
				PADDQ(x2, a[n])
			}
		}
		ADDQ(U32(0x10*0x40), xinput.Base)
		SUBQ(U32(0x10*0x40), plen)

		// scramble xacc
		for n := 0; n < 4; n++ {
			x0 := XMM()
			MOVOU(a[n], x0)
			// xacc[i] ^= (xacc[i] >> 47)
			PSRLQ(Imm(0x2f), a[n])
			PXOR(x0, a[n])
			// xacc[i] ^= xsecret
			PXOR(xsecret.Offset(8*0x10+n*0x10), a[n])
			PSHUFD(Imm(0xf5), a[n], x0)
			PMULULQ(prime, x0)
			// xacc[i] *= prime;
			PSLLQ(Imm(0x20), x0)
			PMULULQ(prime, a[n])
			PADDQ(x0, a[n])
		}
		JMP(LabelRef("accumBlock"))
	}

	// last partial block (64 bytes)
	Label("accumStripe")
	{
		CMPQ(plen, Imm(0x40))
		JLE(LabelRef("accumLastStripe"))

		for n := 0; n < 4; n++ {
			x0, x1, x2 := XMM(), XMM(), XMM()
			// data_vec    = xinput[i]
			MOVOU(xinput.Offset(n*0x10), x0)
			// key_vec     = xsecret[i]
			MOVOU(skey.Offset(n*0x10), x1)
			// data_key    = data_vec ^ key_vec
			PXOR(x0, x1)
			// product     = (data_key & 0xffffffff) * (data_key_lo & 0xffffffff)
			PSHUFD(Imm(0x31), x1, x2)
			PMULULQ(x1, x2)
			// xacc[i] += swap(data_vec)
			PSHUFD(Imm(0x4e), x0, x0)
			PADDQ(x0, a[n])
			// xacc[i] += product
			PADDQ(x2, a[n])
		}
		ADDQ(U32(0x40), xinput.Base)
		SUBQ(U32(0x40), plen)
		ADDQ(U32(8), skey.Base)

		JMP(LabelRef("accumStripe"))
	}

	// last stripe 16 bytes
	Label("accumLastStripe")
	{
		CMPQ(plen, Imm(0))
		JE(LabelRef("return"))

		SUBQ(Imm(0x40), xinput.Base)
		ADDQ(plen, xinput.Base)

		for n := 0; n < 4; n++ {
			x0, x1, x2 := XMM(), XMM(), XMM()
			// data_vec    = xinput[i]
			MOVOU(xinput.Offset(n*0x10), x0)
			// key_vec     = xsecret[i]
			MOVOU(xsecret.Offset(121+n*0x10), x1)
			// data_key    = data_vec ^ key_vec
			PXOR(x0, x1)
			// product     = (data_key & 0xffffffff) * (data_key_lo & 0xffffffff)
			PSHUFD(Imm(0x31), x1, x2)
			PMULULQ(x1, x2)
			// xacc[i] += swap(data_vec)
			PSHUFD(Imm(0x4e), x0, x0)
			PADDQ(x0, a[n])
			// xacc[i] += product
			PADDQ(x2, a[n])
		}
	}

	Label("return")
	{
		MOVOU(a[0], acc.Offset(0x00))
		MOVOU(a[1], acc.Offset(0x10))
		MOVOU(a[2], acc.Offset(0x20))
		MOVOU(a[3], acc.Offset(0x30))
		RET()
	}

	Generate()
}
