// Copyright 2024 ByteDance Inc.
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

#include "textflag.h"
#include "funcdata.h"

TEXT ·compareAndSwapUint128(SB), NOSPLIT, $0-41
	MOVD	addr+0(FP), R0
	MOVD	old1+8(FP), R2
	MOVD	old2+16(FP), R3
	MOVD	new1+24(FP), R4
	MOVD	new2+32(FP), R5
	MOVBU	·arm64HasAtomics+0(SB), R1
	CBZ 	R1, load_store_loop
	MOVD	R2, R6
	MOVD	R3, R7
	CASPD	(R2, R3), (R0), (R4, R5)
	CMP 	R2, R6
	BNE 	ok
	CMP 	R3, R7
	CSET	EQ, R0
	MOVB	R0, ret+40(FP)
	RET
load_store_loop:
	LDAXP	(R0), (R6, R7)
	CMP 	R2, R6
	BNE 	ok
	CMP 	R3, R7
	BNE 	ok
	STLXP	(R4, R5), (R0), R6
	CBNZ	R6, load_store_loop
ok:
	CSET	EQ, R0
	MOVB	R0, ret+40(FP)
	RET

TEXT ·loadUint128(SB),NOSPLIT,$0-24
	MOVD	ptr+0(FP), R0
	LDAXP	(R0), (R0, R1)
	MOVD	R0, ret+8(FP)
	MOVD	R1, ret+16(FP)
	RET

TEXT ·loadSCQNodeUint64(SB),NOSPLIT,$0
	MOVD	ptr+0(FP), R0
	LDAXP	(R0), (R0, R1)
	MOVD	R0, ret+8(FP)
	MOVD	R1, ret+16(FP)
	RET

TEXT ·loadSCQNodePointer(SB),NOSPLIT,$0
	MOVD	ptr+0(FP), R0
	LDAXP	(R0), (R0, R1)
	MOVD	R0, ret+8(FP)
	MOVD	R1, ret+16(FP)
	RET

TEXT ·atomicTestAndSetFirstBit(SB),NOSPLIT,$0
	MOVD	addr+0(FP), R0
load_store_loop:
	LDAXR	(R0), R1
	ORR 	$(1<<63), R1, R1
	STLXR	R1, (R0), R2
	CBNZ	R2, load_store_loop
	MOVD	R1, val+8(FP)
	RET


TEXT ·atomicTestAndSetSecondBit(SB),NOSPLIT,$0
	MOVD	addr+0(FP), R0
load_store_loop:
	LDAXR	(R0), R1
	ORR 	$(1<<62), R1, R1
	STLXR	R1, (R0), R2
	CBNZ	R2, load_store_loop
	MOVD	R1, val+8(FP)
	RET

TEXT ·resetNode(SB),NOSPLIT,$0
	MOVD	addr+0(FP), R0
	MOVD	$0, 8(R0)
load_store_loop:
	LDAXR	(R0), R1
	ORR 	$(1<<62), R1, R1
	STLXR	R1, (R0), R2
	CBNZ	R2, load_store_loop
	RET

TEXT ·runtimeEnableWriteBarrier(SB),NOSPLIT,$0
	MOVW	runtime·writeBarrier(SB), R0
	MOVB	R0, ret+0(FP)
	RET
