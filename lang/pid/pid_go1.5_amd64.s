// Copyright 2016 Peter Mattis.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.

// This file has been modified by Cholerae Hu.
// Copyright 2022 Cholerae Hu.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License. See the AUTHORS file
// for names of contributors.

// This file may have been modified by ByteDance Inc. ("ByteDance Modifications"). All ByteDance Modifications are Copyright 2022 ByteDance Inc.

// Assembly to mimic runtime.getg.

// +build amd64 amd64p32
// +build gc,go1.5

#include "go_asm.h"
#include "textflag.h"

// func getPid() int64
TEXT ·getPid(SB),NOSPLIT,$0-8
	MOVQ (TLS), R14
	MOVQ g_m(R14), R13
	MOVQ m_p(R13), R14
	MOVL p_id(R14), R13
	MOVQ R13, ret+0(FP)
	RET
