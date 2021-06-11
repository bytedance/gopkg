// +build amd64,!gccgo,!appengine

#include "textflag.h"

TEXT ·compareAndSwapUint128(SB),NOSPLIT,$0
	MOVQ addr+0(FP), R8
	MOVQ old+8(FP), AX
	MOVQ old+16(FP), DX
	MOVQ new+24(FP), BX
	MOVQ new+32(FP), CX
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
	MOVQ AX, val+8(FP)
	MOVQ DX, val+16(FP)
	RET

TEXT ·loadSCQNodeUint64(SB),NOSPLIT,$0
	JMP ·loadUint128(SB)

TEXT ·loadSCQNodePointer(SB),NOSPLIT,$0
	JMP ·loadUint128(SB)

TEXT ·compareAndSwapSCQNodePointer(SB),NOSPLIT,$0
	JMP ·compareAndSwapUint128(SB)

TEXT ·compareAndSwapSCQNodeUint64(SB),NOSPLIT,$0
	JMP ·compareAndSwapUint128(SB)

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
