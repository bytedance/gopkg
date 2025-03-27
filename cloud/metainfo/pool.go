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

import "sync"

var tmpnodePool = sync.Pool{}

type tmpnode struct {
	persistent []kv
	transient  []kv
}

const (
	// tmpnodeDefaultBufferSize is the default cap size of slice used by tmpnode
	tmpnodeDefaultBufferSize = 32

	// tmpnodeCopyThresholdSize is the threshold of copying data from `tmpnode` to `node`
	//
	// if len(persistent) or len(transient) > the value, the new `node` will borrow refs from `tmpnode`
	// see `Node` method for details.
	tmpnodeCopyThresholdSize = 0.75 * tmpnodeDefaultBufferSize
)

// Node copies data from persistent or transient, and returns a new `*node` for context
func (n *tmpnode) Node() *node {
	ret := &node{}

	// in case we only have a few kvs, copy cost is low
	if sz := n.Size(); sz <= 2*tmpnodeCopyThresholdSize {
		kvs := make([]kv, n.Size()) // alloc one, and used by persistent & transient
		sz = copy(kvs, n.persistent)
		ret.persistent = kvs[:sz:sz]
		kvs = kvs[sz:]
		sz = copy(kvs, n.transient)
		ret.transient = kvs
		return ret
	}

	// if size of the slice has items more than `tmpnodeCopyThresholdSize`
	// considering reusing the slice instead of copying data.
	if sz := len(n.persistent); sz > tmpnodeCopyThresholdSize {
		ret.persistent = n.persistent
		n.persistent = nil
	} else if sz > 0 {
		ret.persistent = append(make([]kv, 0, sz), n.persistent...)
	}

	if sz := len(n.transient); sz > tmpnodeCopyThresholdSize {
		ret.transient = n.transient
		n.transient = nil
	} else if sz > 0 {
		ret.transient = append(make([]kv, 0, sz), n.transient...)
	}
	return ret
}

func (n *tmpnode) Size() int {
	return len(n.persistent) + len(n.transient)
}

func (n *tmpnode) Reset() {
	if n.persistent == nil && n.transient == nil {
		// in case both nil,
		// alloc one slice and used by persistent & transient
		// one less allocation
		kvs := make([]kv, 2*tmpnodeDefaultBufferSize)
		n.persistent = kvs[:tmpnodeDefaultBufferSize][:0:tmpnodeDefaultBufferSize]
		n.transient = kvs[tmpnodeDefaultBufferSize:][:0:tmpnodeDefaultBufferSize]
		return
	}

	if n.persistent == nil { // set to nil in Node() if it's large
		n.persistent = make([]kv, 0, tmpnodeDefaultBufferSize)
	} else {
		n.persistent = n.persistent[:0]
	}
	if n.transient == nil { // set to nil in Node() if it's large
		n.transient = make([]kv, 0, tmpnodeDefaultBufferSize)
	} else {
		n.transient = n.transient[:0]
	}
}
