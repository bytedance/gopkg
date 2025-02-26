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

package metainfo

import "testing"

func assert(t *testing.T, cond bool, val ...interface{}) {
	t.Helper()
	if !cond {
		if len(val) > 0 {
			t.Fatal(val...)
		} else {
			t.Fatal("assertion failed")
		}
	}
}

func TestTmpNode(t *testing.T) {
	n := tmpnode{}

	// Test Reset, case both nil
	n.persistent, n.transient = nil, nil
	n.Reset()
	assert(t, len(n.persistent) == 0)
	assert(t, cap(n.persistent) == tmpnodeDefaultBufferSize)
	assert(t, len(n.transient) == 0)
	assert(t, cap(n.transient) == tmpnodeDefaultBufferSize)

	// Test Reset, n.persistent = nil
	n.persistent = nil
	n.transient = make([]kv, 7, 7)
	n.Reset()
	assert(t, len(n.persistent) == 0)
	assert(t, cap(n.persistent) == tmpnodeDefaultBufferSize)
	assert(t, len(n.transient) == 0)
	assert(t, cap(n.transient) == 7)

	// Test Reset,  n.transient = nil
	n.persistent = make([]kv, 7, 7)
	n.transient = nil
	n.Reset()
	assert(t, len(n.persistent) == 0)
	assert(t, cap(n.persistent) == 7)
	assert(t, len(n.transient) == 0)
	assert(t, cap(n.transient) == tmpnodeDefaultBufferSize)

	// Test Node(), <= 2*tmpnodeCopyThresholdSize
	n.persistent = []kv{{"k1", "v1"}}
	n.transient = []kv{{"k2", "v2"}}
	newn := n.Node()
	assert(t, len(n.persistent) == 1)
	assert(t, cap(n.persistent) == 1)
	assert(t, len(n.transient) == 1)
	assert(t, cap(n.transient) == 1)
	assert(t, n.persistent[0] == newn.persistent[0])
	assert(t, n.transient[0] == newn.transient[0])

	// Test Node(), > tmpnodeCopyThresholdSize
	sz := int(2 * tmpnodeCopyThresholdSize)
	n.persistent = make([]kv, sz, sz)
	n.persistent[0] = kv{"k3", "v3"}
	n.transient = []kv{{"k4", "v4"}}
	newn = n.Node()
	assert(t, n.persistent == nil)
	assert(t, len(n.transient) == 1)
	assert(t, len(newn.persistent) == sz)
	assert(t, newn.persistent[0] == kv{"k3", "v3"})
	assert(t, len(newn.transient) == 1)
	assert(t, n.transient[0] == newn.transient[0])

	n.persistent = []kv{{"k3", "v3"}}
	n.transient = make([]kv, sz, sz)
	n.transient[0] = kv{"k4", "v4"}
	newn = n.Node()
	assert(t, len(n.persistent) == 1)
	assert(t, n.transient == nil)
	assert(t, len(newn.persistent) == 1)
	assert(t, newn.persistent[0] == n.persistent[0])
	assert(t, len(newn.transient) == sz)
	assert(t, newn.transient[0] == kv{"k4", "v4"})

}
