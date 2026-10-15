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

package xxhash3

import (
	"bytes"
	"testing"
	"unsafe"
)

// TestXsecretAligned guards against a regression of
// https://github.com/mmcloughlin/avo/issues/497: the SSE2 accumulator uses a
// legacy PXOR instruction with a direct memory operand into xsecret, which
// the CPU requires to be 16-byte aligned. Before this fix, xsecret's
// alignment was down to incidental allocator placement of a plain
// [192]uint8, so this could pass or fault depending on Go version and luck.
func TestXsecretAligned(t *testing.T) {
	if addr := uintptr(xsecret); addr%xsecretAlign != 0 {
		t.Fatalf("xsecret = %#x is not %d-byte aligned", addr, xsecretAlign)
	}
}

func TestXsecretContents(t *testing.T) {
	got := unsafe.Slice((*byte)(xsecret), len(xsecretSrc))
	if !bytes.Equal(got, xsecretSrc[:]) {
		t.Fatalf("aligned copy of xsecret does not match xsecretSrc\ngot:  %x\nwant: %x", got, xsecretSrc[:])
	}
}
