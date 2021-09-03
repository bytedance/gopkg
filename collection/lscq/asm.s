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

// +build amd64,!gccgo,!appengine

#include "textflag.h"

TEXT ·compareAndSwapUint128(SB),NOSPLIT,$0
	MOVQ addr+0(FP), R8
	MOVQ old1+8(FP), AX
	MOVQ old2+16(FP), DX
	MOVQ new1+24(FP), BX
	MOVQ new2+32(FP), CX
	LOCK
	CMPXCHG16B (R8)
	SETEQ swapped+40(FP)
	RET

TEXT ·loadUint128(SB),NOSPLIT,$0
	MOVQ addr+0(FP), R8
	XORQ AX, AX
	XORQ DX, DX
	XORQ BX, BX
	XORQ CX, CX
	LOCK
	CMPXCHG16B (R8)
	MOVQ AX, val_0+8(FP)
	MOVQ DX, val_1+16(FP)
	RET

TEXT ·loadSCQNodeUint64(SB),NOSPLIT,$0
	JMP ·loadUint128(SB)

TEXT ·loadSCQNodePointer(SB),NOSPLIT,$0
	JMP ·loadUint128(SB)

TEXT ·atomicTestAndSetFirstBit(SB),NOSPLIT,$0
	MOVQ addr+0(FP), DX
	LOCK
	BTSQ $63,(DX)
	MOVQ AX, val+8(FP)
	RET

TEXT ·atomicTestAndSetSecondBit(SB),NOSPLIT,$0
	MOVQ addr+0(FP), DX
	LOCK
	BTSQ $62,(DX)
	MOVQ AX, val+8(FP)
	RET

TEXT ·resetNode(SB),NOSPLIT,$0
	MOVQ addr+0(FP), DX
	MOVQ $0, 8(DX)
	LOCK
	BTSQ $62, (DX)
	RET

TEXT ·runtimeEnableWriteBarrier(SB),NOSPLIT,$0
	MOVL runtime·writeBarrier(SB), AX
	MOVB AX, ret+0(FP)
	RET
