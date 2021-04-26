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

#include "textflag.h"

#define SYS_clock_gettime	228

// func now() (sec int64, nsec int32, mono int64)
TEXT ·now(SB),NOSPLIT,$16-24
	MOVQ	SP, R12	// Save old SP; R12 unchanged by C code.

noswitch:
	SUBQ	$16, SP		// Space for results
	ANDQ	$~15, SP	// Align for C code

    // call REALTIME first
	MOVL	$5, DI // CLOCK_REALTIME_COARSE
	LEAQ	0(SP), SI
	MOVQ	runtime·vdsoClockgettimeSym(SB), AX
	CMPQ	AX, $0
	JEQ	fallback
	CALL	AX
	// save REALTIME ret
	MOVQ	0(SP), AX	// sec
    MOVQ	8(SP), DX	// nsec
    MOVQ    SP, R11     // save current SP
    MOVQ	R12, SP		// Restore real SP
	MOVQ	AX, sec+0(FP)	// sec
    MOVL	DX, nsec+8(FP)	// nsec
    MOVQ    R11, SP     // Restore current SP

    // call MONOTONIC time
    MOVL	$6, DI // CLOCK_MONOTONIC_COARSE
    LEAQ	0(SP), SI
    MOVQ	runtime·vdsoClockgettimeSym(SB), AX
    CALL	AX
ret:
	MOVQ	0(SP), AX	// sec
	MOVQ	8(SP), DX	// nsec
	MOVQ	R12, SP		// Restore real SP
	// sec is in AX, nsec in DX
	// return nsec in AX
	IMULQ	$1000000000, AX
	ADDQ	DX, AX
	MOVQ	AX, mono+16(FP)
	RET
fallback:
	MOVQ	$SYS_clock_gettime, AX
	SYSCALL
	// save REALTIME ret
    MOVQ	0(SP), AX	// sec
    MOVQ	8(SP), DX	// nsec
    MOVQ    SP, R11     // save current SP
    MOVQ	R12, SP		// Restore real SP
    MOVQ	AX, sec+0(FP)	// sec
    MOVL	DX, nsec+8(FP)	// nsec
    MOVQ    R11, SP     // Restore current SP

    MOVL	$6, DI // CLOCK_MONOTONIC_COARSE
    LEAQ	0(SP), SI
    MOVQ	$SYS_clock_gettime, AX
    SYSCALL
	JMP	ret

TEXT ·startNano(SB),NOSPLIT,$0-8
	MOVQ	time·startNano(SB), AX
	MOVQ    AX, ret+0(FP)
	RET
