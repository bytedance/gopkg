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

#include "go_asm.h"
#include "textflag.h"

// go_tls.h
#define	get_tls(r)	MOVQ TLS, r
#define	g(r)	0(r)(TLS*1)

#define SYS_clock_gettime	228

// func walltime1() (sec int64, nsec int32)
// non-zero frame-size means bp is saved and restored
TEXT ·walltime1(SB),NOSPLIT,$16-12
	// We don't know how much stack space the VDSO code will need,
	// so switch to g0.
	// In particular, a kernel configured with CONFIG_OPTIMIZE_INLINING=n
	// and hardening can use a full page of stack space in gettime_sym
	// due to stack probes inserted to avoid stack/heap collisions.
	// See issue #20427.

	MOVQ	SP, R12	// Save old SP; R12 unchanged by C code.

	get_tls(CX)
	MOVQ	g(CX), AX
	MOVQ	g_m(AX), BX // BX unchanged by C code.

	// Set vdsoPC and vdsoSP for SIGPROF traceback.
	// Save the old values on stack and restore them on exit,
	// so this function is reentrant.
	MOVQ	m_vdsoPC(BX), CX
	MOVQ	m_vdsoSP(BX), DX
	MOVQ	CX, 0(SP)
	MOVQ	DX, 8(SP)

	LEAQ	sec+0(FP), DX
	MOVQ	-8(DX), CX
	MOVQ	CX, m_vdsoPC(BX)
	MOVQ	DX, m_vdsoSP(BX)

	CMPQ	AX, m_curg(BX)	// Only switch if on curg.
	JNE	noswitch

	MOVQ	m_g0(BX), DX
	MOVQ	(g_sched+gobuf_sp)(DX), SP	// Set SP to g0 stack

noswitch:
	SUBQ	$16, SP		// Space for results
	ANDQ	$~15, SP	// Align for C code

	MOVL	$5, DI // CLOCK_REALTIME_COARSE
	LEAQ	0(SP), SI
	MOVQ	runtime·vdsoClockgettimeSym(SB), AX
	CMPQ	AX, $0
	JEQ	fallback
	CALL	AX
ret:
	MOVQ	0(SP), AX	// sec
	MOVQ	8(SP), DX	// nsec
	MOVQ	R12, SP		// Restore real SP
	// Restore vdsoPC, vdsoSP
	// We don't worry about being signaled between the two stores.
	// If we are not in a signal handler, we'll restore vdsoSP to 0,
	// and no one will care about vdsoPC. If we are in a signal handler,
	// we cannot receive another signal.
	MOVQ	8(SP), CX
	MOVQ	CX, m_vdsoSP(BX)
	MOVQ	0(SP), CX
	MOVQ	CX, m_vdsoPC(BX)
	MOVQ	AX, sec+0(FP)
	MOVL	DX, nsec+8(FP)
	RET
fallback:
	MOVQ	$SYS_clock_gettime, AX
	SYSCALL
	JMP ret

// func nanotime1() int64
TEXT ·nanotime1(SB),NOSPLIT,$16-8
	// Switch to g0 stack. See comment above in runtime·walltime.

	MOVQ	SP, R12	// Save old SP; R12 unchanged by C code.

	get_tls(CX)
	MOVQ	g(CX), AX
	MOVQ	g_m(AX), BX // BX unchanged by C code.

	// Set vdsoPC and vdsoSP for SIGPROF traceback.
	// Save the old values on stack and restore them on exit,
	// so this function is reentrant.
	MOVQ	m_vdsoPC(BX), CX
	MOVQ	m_vdsoSP(BX), DX
	MOVQ	CX, 0(SP)
	MOVQ	DX, 8(SP)

	LEAQ	ret+0(FP), DX
	MOVQ	-8(DX), CX
	MOVQ	CX, m_vdsoPC(BX)
	MOVQ	DX, m_vdsoSP(BX)

	CMPQ	AX, m_curg(BX)	// Only switch if on curg.
	JNE	noswitch

	MOVQ	m_g0(BX), DX
	MOVQ	(g_sched+gobuf_sp)(DX), SP	// Set SP to g0 stack

noswitch:
	SUBQ	$16, SP		// Space for results
	ANDQ	$~15, SP	// Align for C code

	MOVL	$6, DI // CLOCK_MONOTONIC_COARSE
	LEAQ	0(SP), SI
	MOVQ	runtime·vdsoClockgettimeSym(SB), AX
	CMPQ	AX, $0
	JEQ	fallback
	CALL	AX
ret:
	MOVQ	0(SP), AX	// sec
	MOVQ	8(SP), DX	// nsec
	MOVQ	R12, SP		// Restore real SP
	// Restore vdsoPC, vdsoSP
	// We don't worry about being signaled between the two stores.
	// If we are not in a signal handler, we'll restore vdsoSP to 0,
	// and no one will care about vdsoPC. If we are in a signal handler,
	// we cannot receive another signal.
	MOVQ	8(SP), CX
	MOVQ	CX, m_vdsoSP(BX)
	MOVQ	0(SP), CX
	MOVQ	CX, m_vdsoPC(BX)
	// sec is in AX, nsec in DX
	// return nsec in AX
	IMULQ	$1000000000, AX
	ADDQ	DX, AX
	MOVQ	AX, ret+0(FP)
	RET
fallback:
	MOVQ	$SYS_clock_gettime, AX
	SYSCALL
	JMP	ret


// func now() (sec int64, nsec int32, mono int64)
TEXT ·now(SB),NOSPLIT,$16-24
	// Switch to g0 stack. See comment above in runtime·walltime.

	MOVQ	SP, R12	// Save old SP; R12 unchanged by C code.

	get_tls(CX)
	MOVQ	g(CX), AX
	MOVQ	g_m(AX), BX // BX unchanged by C code.

	// Set vdsoPC and vdsoSP for SIGPROF traceback.
	// Save the old values on stack and restore them on exit,
	// so this function is reentrant.
	MOVQ	m_vdsoPC(BX), CX
	MOVQ	m_vdsoSP(BX), DX
	MOVQ	CX, 0(SP)
	MOVQ	DX, 8(SP)

	LEAQ	sec+0(FP), DX // not really FP, just for make vet happy
	MOVQ	-8(DX), CX
	MOVQ	CX, m_vdsoPC(BX)
	MOVQ	DX, m_vdsoSP(BX)

	CMPQ	AX, m_curg(BX)	// Only switch if on curg.
	JNE	noswitch

	MOVQ	m_g0(BX), DX
	MOVQ	(g_sched+gobuf_sp)(DX), SP	// Set SP to g0 stack

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
	// Restore vdsoPC, vdsoSP
	// We don't worry about being signaled between the two stores.
	// If we are not in a signal handler, we'll restore vdsoSP to 0,
	// and no one will care about vdsoPC. If we are in a signal handler,
	// we cannot receive another signal.
	MOVQ	8(SP), CX
	MOVQ	CX, m_vdsoSP(BX)
	MOVQ	0(SP), CX
	MOVQ	CX, m_vdsoPC(BX)
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
