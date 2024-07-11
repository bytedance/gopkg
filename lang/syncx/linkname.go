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

//go:build !race
// +build !race

package syncx

import (
	_ "sync"
	_ "unsafe"
)

//go:noescape
//go:linkname runtime_registerPoolCleanup sync.runtime_registerPoolCleanup
func runtime_registerPoolCleanup(cleanup func())

//go:noescape
//go:linkname runtime_poolCleanup sync.poolCleanup
func runtime_poolCleanup()

//go:noescape
//go:linkname runtime_Semacquire sync.runtime_Semacquire
func runtime_Semacquire(s *uint32)

//go:noescape
//go:linkname runtime_Semrelease sync.runtime_Semrelease
func runtime_Semrelease(s *uint32, handoff bool, skipframes int)
