#include "textflag.h"

// func lzcnt(x int) int
TEXT ·lzcnt(SB), NOSPLIT, $0-16
    LZCNTQ x+0(FP), AX
    MOVQ AX, ret+8(FP)
    RET

// func bsr(x int) int
TEXT ·bsr(SB), NOSPLIT, $0-16
    BSRQ x+0(FP), AX
    MOVQ AX, ret+8(FP)
    RET
