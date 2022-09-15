// Copyright 2022 Loongson Inc.
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

// +build loong64,!gccgo,!appengine

#include "textflag.h"

//func loadUint128(addr *uint128) (ret uint128)
TEXT ·loadUint128Asm(SB),NOSPLIT,$0-24
	MOVV    addr+0(FP), R4
	DBAR
	MOVV	0(R4), R5
	MOVV	8(R4), R6
	DBAR
	MOVV	R5, ret+8(FP)
	MOVV	R6, ret+16(FP)
	DBAR
	RET

// func loadSCQNodeUint64(addr unsafe.Pointer)(val scqNodeUint64)
TEXT ·loadSCQNodeUint64(SB),NOSPLIT,$0
	JMP ·loadUint128(SB)

// func loadSCQNodePointer(addr unsafe.Pointer)(val scqNodePointer)
TEXT ·loadSCQNodePointer(SB),NOSPLIT,$0
	JMP ·loadUint128(SB)

// func atomicTestAndSetFirstBit(addr *uint64)(val uint64)
TEXT ·atomicTestAndSetFirstBit(SB),NOSPLIT,$0
	MOVV	addr+0(FP), R4
	MOVV	$(1<<63), R6
	DBAR
testAndSetFirstBit_again:
	LLV	(R4), R5
	OR	R6, R5, R5
	MOVV	R5, R7
	SCV	R5, (R4)
	BEQ	R5, testAndSetFirstBit_again
	DBAR
	MOVV	R7, val+8(FP)
	RET

// func atomicTestAndSetSecondBit(addr *uint64)(val uint64)
TEXT ·atomicTestAndSetSecondBit(SB),NOSPLIT,$0
	MOVV	addr+0(FP), R4
	MOVV	$(1<<62), R6
	DBAR
testAndSetSecBit_again:
	LLV	(R4), R5
	OR	R6, R5, R5
	MOVV	R5, R7
	SCV	R5, (R4)
	BEQ	R5, testAndSetSecBit_again
	DBAR
	MOVV	R7, val+8(FP)
	RET

// func resetNode(addr unsafe.Pointer)
TEXT ·resetNode(SB),NOSPLIT,$0
        MOVV	addr+0(FP), R4
        MOVV	R0, 8(R4)
	MOVV	$(1<<62), R6
        DBAR
resetNode_again:
	LLV	(R4), R5
	OR	R6, R5, R5
	SCV	R5, (R4)
	BEQ	R5, resetNode_again
	DBAR
	RET

// func runtimeEnableWriteBarrier() bool
TEXT ·runtimeEnableWriteBarrier(SB),NOSPLIT,$0
	MOVW runtime·writeBarrier(SB), R0
	DBAR
	MOVB R0, ret+0(FP)
	RET
