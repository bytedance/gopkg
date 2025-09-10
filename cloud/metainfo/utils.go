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

	p := poolKVLoader.Get().(*kvLoader)
	defer poolKVLoader.Put(p)
	if p.Load(m) == 0 {
		//  no new kv added?
		return ctx
	}
	return withNode(ctx, p.Merge(nd))
}

func newCtxFromMap(ctx context.Context, m map[string]string) context.Context {

	loader := poolKVLoader.Get().(*kvLoader)
	defer poolKVLoader.Put(loader)
	// return original ctx if no valid key in map
	if loader.Load(m) == 0 {
		return ctx
	}
	return withNode(ctx, loader.Node())
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

var poolKVLoader = sync.Pool{
	New: func() interface{} {
		p := &kvLoader{}
		p.dup = make(map[string]bool, 8)
		p.persistent = make([]kv, 0, 8)
		p.transient = make([]kv, 0, 8)
		p.stale = make([]kv, 0, 8)
		return p
	},
}

type kvLoader struct {
	dup        map[string]bool
	persistent []kv // PrefixPersistent
	transient  []kv // PrefixTransient
	stale      []kv // PrefixTransientUpstream
}

func (p *kvLoader) Node() *node {
	ret := &node{}
	kvs := make([]kv, len(p.persistent)+len(p.transient)+len(p.stale))
	if n := len(p.persistent); n != 0 {
		copy(kvs, p.persistent)
		ret.persistent = kvs[:n:n]
		kvs = kvs[n:]
	}
	if n := len(p.transient); n != 0 {
		copy(kvs, p.transient)
		ret.transient = kvs[:n:n]
		kvs = kvs[n:]
	}
	if len(p.stale) != 0 {
		copy(kvs, p.stale)
		ret.stale = kvs
	}
	return ret
}

func (p *kvLoader) String() string {
	return fmt.Sprintf("persistent:%v, transient:%v, stale:%v",
		p.persistent, p.transient, p.stale)
}

func (p *kvLoader) Load(m map[string]string) int {
	p.persistent = p.persistent[:0]
	p.transient = p.transient[:0]
	p.stale = p.stale[:0]
	for k, v := range m {
		klen := len(k)
		if klen == 0 || len(v) == 0 {
			continue
		}

		// Check for PrefixTransientUpstream first (longest prefix)
		if klen >= lenPTU && k[:lenPTU] == PrefixTransientUpstream {
			if klen > lenPTU { // skip empty key
				p.stale = append(p.stale, kv{key: k[lenPTU:], val: v})
			}
		} else if klen >= lenPT && k[:lenPT] == PrefixTransient {
			if klen > lenPT {
				p.transient = append(p.transient, kv{key: k[lenPT:], val: v})
			}
		} else if klen >= lenPP && k[:lenPP] == PrefixPersistent {
			if klen > lenPP {
				p.persistent = append(p.persistent, kv{key: k[lenPP:], val: v})
			}
		}
	}
	return len(p.stale) + len(p.transient) + len(p.persistent)
}

func mergekv(dup map[string]bool, newkvs, oldkvs []kv) []kv {
	// will be optimized by compiler -> mapclear
	for k := range dup {
		delete(dup, k)
	}
	for i := range newkvs {
		dup[newkvs[i].key] = true
	}
	for j := range oldkvs {
		if !dup[oldkvs[j].key] {
			newkvs = append(newkvs, oldkvs[j])
		}
	}
	return newkvs
}

func (p *kvLoader) Merge(old *node) *node {
	if len(p.persistent) != 0 {
		p.persistent = mergekv(p.dup, p.persistent, old.persistent)
	}
	if len(p.transient) != 0 {
		p.transient = mergekv(p.dup, p.transient, old.transient)
	}
	if len(p.stale) != 0 {
		p.stale = mergekv(p.dup, p.stale, old.stale)
	}
	ret := p.Node()
	// reuse the old node if nothing to merge
	if len(ret.persistent) == 0 {
		ret.persistent = old.persistent
	}
	if len(ret.transient) == 0 {
		ret.transient = old.transient
	}
	if len(ret.stale) == 0 {
		ret.stale = old.stale
	}
	return ret
}
