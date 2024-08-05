// Copyright 2023 ByteDance Inc.
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

package metainfo

import "sync"

type kvstore map[string]string

var kvpool sync.Pool

func newKVStore(size ...int) kvstore {
	kvs := kvpool.Get()
	if kvs == nil {
		if len(size) > 0 {
			return make(kvstore, size[0])
		}
		return make(kvstore)
	}
	return kvs.(kvstore)
}

func (store kvstore) size() int {
	return len(store)
}

func (store kvstore) recycle() {
	/*
			for k := range m {
				delete(m, k)
			}
		  ==>
			LEAQ    type.map[string]int(SB), AX
			MOVQ    AX, (SP)
			MOVQ    "".m(SB), AX
			MOVQ    AX, 8(SP)
			PCDATA  $1, $0
			CALL    runtime.mapclear(SB)
	*/
	for key := range store {
		delete(store, key)
	}
	kvpool.Put(store)
}
