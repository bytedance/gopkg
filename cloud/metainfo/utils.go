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
	"strings"
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
	// inherit from node
	mapSize := len(m)
	persistent := newKVStore(mapSize)
	transient := newKVStore(mapSize)
	stale := newKVStore(mapSize)
	sliceToMap(nd.persistent, persistent)
	sliceToMap(nd.transient, transient)
	sliceToMap(nd.stale, stale)

	// insert new kvs from m to node
	for k, v := range m {
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		switch {
		case strings.HasPrefix(k, PrefixTransientUpstream):
			if len(k) > lenPTU { // do not move this condition to the case statement to prevent a PTU matches PT
				stale[k[lenPTU:]] = v
			}
		case strings.HasPrefix(k, PrefixTransient):
			if len(k) > lenPT {
				transient[k[lenPT:]] = v
			}
		case strings.HasPrefix(k, PrefixPersistent):
			if len(k) > lenPP {
				persistent[k[lenPP:]] = v
			}
		}
	}

	// return original ctx if no invalid key in map
	if (persistent.size() + transient.size() + stale.size()) == 0 {
		return ctx
	}

	// make new node, and transfer map to list
	nd = newNodeFromMaps(persistent, transient, stale)
	persistent.recycle()
	transient.recycle()
	stale.recycle()
	return withNode(ctx, nd)
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
