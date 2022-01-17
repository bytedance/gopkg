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

import "flag"

var (
	avx = flag.Bool("avx2", false, "avx2")
	sse = flag.Bool("sse2", false, "sse2")
)

func main() {
	flag.Parse()

	if *avx {
		AVX2()
	} else if *sse {
		SSE2()
	}
}
