// Copyright 2019 Cholerae Hu.
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

//go:build (amd64 || amd64p32 || arm64) && !windows && gc && go1.5
// +build amd64 amd64p32 arm64
// +build !windows
// +build gc
// +build go1.5

package goid

//go:nosplit
func getPid() uintptr

//go:nosplit
func GetPid() int {
	return int(getPid())
}
