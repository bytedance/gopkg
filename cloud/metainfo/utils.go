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

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// HasMetaInfo detects whether the given context contains metainfo.
func HasMetaInfo(ctx context.Context) bool {
	return getNode(ctx) != nil
}

// SetMetaInfoFromMap retrieves metainfo key-value pairs from the given map and sets then into the context.
// Only those keys with prefixes defined in this module would be used.
// If the context has been carrying metanifo pairs, they will be merged as a basis.
func SetMetaInfoFromMap(ctx context.Context, m map[string]string) context.Context {
	if ctx == nil || len(m) == 0 {
		return ctx
	}

	nd := getNode(ctx)
	if nd == nil || nd.size() == 0 {
		// fast path
		return newCtxFromMap(ctx, m)
	}

	p := poolKVMerge.Get().(*kvMerge)
	defer poolKVMerge.Put(p)
	if p.Load(m) == 0 {
		//  no new kv added?
		return ctx
	}
	return withNode(ctx, p.Merge(nd))
}

func newCtxFromMap(ctx context.Context, m map[string]string) context.Context {
	// make new node
	mapSize := len(m)
	nd := &node{
		persistent: make([]kv, 0, mapSize),
		transient:  make([]kv, 0, mapSize),
		stale:      make([]kv, 0, mapSize),
	}

	// insert new kvs from m to node
	for k, v := range m {
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		switch {
		case strings.HasPrefix(k, PrefixTransientUpstream):
			if len(k) > lenPTU { // do not move this condition to the case statement to prevent a PTU matches PT
				nd.stale = append(nd.stale, kv{key: k[lenPTU:], val: v})
			}
		case strings.HasPrefix(k, PrefixTransient):
			if len(k) > lenPT {
				nd.transient = append(nd.transient, kv{key: k[lenPT:], val: v})
			}
		case strings.HasPrefix(k, PrefixPersistent):
			if len(k) > lenPP {
				nd.persistent = append(nd.persistent, kv{key: k[lenPP:], val: v})
			}
		}
	}

	// return original ctx if no invalid key in map
	if nd.size() == 0 {
		return ctx
	}
	return withNode(ctx, nd)
}

// SaveMetaInfoToMap set key-value pairs from ctx to m while filtering out transient-upstream data.
func SaveMetaInfoToMap(ctx context.Context, m map[string]string) {
	if ctx == nil || m == nil {
		return
	}
	ctx = TransferForward(ctx)
	if n := getNode(ctx); n != nil {
		for _, kv := range n.stale {
			m[PrefixTransient+kv.key] = kv.val
		}
		for _, kv := range n.transient {
			m[PrefixTransient+kv.key] = kv.val
		}
		for _, kv := range n.persistent {
			m[PrefixPersistent+kv.key] = kv.val
		}
	}
}

// sliceToMap converts a kv slice to map.
func sliceToMap(slice []kv, kvs kvstore) {
	if len(slice) == 0 {
		return
	}
	for _, kv := range slice {
		kvs[kv.key] = kv.val
	}
}

var poolKVMerge = sync.Pool{
	New: func() interface{} {
		p := &kvMerge{}
		p.dup = make(map[string]bool, 8)
		p.persistent = make([]kv, 0, 8)
		p.transient = make([]kv, 0, 8)
		p.stale = make([]kv, 0, 8)
		return p
	},
}

type kvMerge struct {
	dup        map[string]bool
	persistent []kv // PrefixPersistent
	transient  []kv // PrefixTransient
	stale      []kv // PrefixTransientUpstream
}

func (p *kvMerge) String() string {
	return fmt.Sprintf("persistent:%v, transient:%v, stale:%v",
		p.persistent, p.transient, p.stale)
}

func (p *kvMerge) resetdup() {
	for k := range p.dup {
		delete(p.dup, k)
	}
}

func (p *kvMerge) Load(m map[string]string) int {
	p.persistent = p.persistent[:0]
	p.transient = p.transient[:0]
	p.stale = p.stale[:0]
	for k, v := range m {
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		switch {
		case strings.HasPrefix(k, PrefixTransient):
			if len(k) <= lenPT {
				continue
			}
			k = k[lenPT:]
			if strings.HasPrefix(k, "UPSTREAM_") { // PrefixTransientUpstream {
				if len(k) > lenU {
					p.stale = append(p.stale, kv{key: k[lenU:], val: v})
				}
			} else {
				p.transient = append(p.transient, kv{key: k, val: v})
			}

		case strings.HasPrefix(k, PrefixPersistent):
			if len(k) > lenPP {
				p.persistent = append(p.persistent, kv{key: k[lenPP:], val: v})
			}
		}
	}
	return len(p.stale) + len(p.transient) + len(p.persistent)
}

func (p *kvMerge) Merge(old *node) *node {
	// this method assumes that
	// keys in old.persistent, old.transient, and old.stale are unique

	if len(p.persistent) != 0 {
		p.resetdup()
		for i := range p.persistent {
			p.dup[p.persistent[i].key] = true
		}
		for j := range old.persistent {
			if !p.dup[old.persistent[j].key] {
				p.persistent = append(p.persistent, old.persistent[j])
			}
		}
	}
	if len(p.transient) != 0 {
		p.resetdup()
		for i := range p.transient {
			p.dup[p.transient[i].key] = true
		}
		for j := range old.transient {
			if !p.dup[old.transient[j].key] {
				p.transient = append(p.transient, old.transient[j])
			}
		}
	}

	if len(p.stale) != 0 {
		p.resetdup()
		for i := range p.stale {
			p.dup[p.stale[i].key] = true
		}
		for j := range old.stale {
			if !p.dup[old.transient[j].key] {
				p.stale = append(p.stale, old.stale[j])
			}
		}
	}

	// copy to ret

	ret := &node{}
	kvs := make([]kv, len(p.persistent)+len(p.transient)+len(p.stale))
	if n := len(p.persistent); n != 0 {
		copy(kvs, p.persistent)
		ret.persistent = kvs[:n:n]
		kvs = kvs[n:]
	} else {
		ret.persistent = old.persistent
	}
	if n := len(p.transient); n != 0 {
		copy(kvs, p.transient)
		ret.transient = kvs[:n:n]
		kvs = kvs[n:]
	} else {
		ret.transient = old.transient
	}
	if len(p.stale) != 0 { // last one, use kvs directly
		copy(kvs, p.stale)
		ret.stale = kvs
	} else {
		ret.stale = old.stale
	}
	return ret
}
