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
	var persistent kvstore
	var transient kvstore
	var stale kvstore

	// inherit from exist ctx node
	mapSize := len(m)
	nd := getNode(ctx)
	inherit := nd != nil
	if inherit {
		// inherit from node
		persistent = newKVStore(mapSize)
		transient = newKVStore(mapSize)
		stale = newKVStore(mapSize)
		sliceToMap(nd.persistent, persistent)
		sliceToMap(nd.transient, transient)
		sliceToMap(nd.stale, stale)
	} else {
		// make new node
		nd = &node{
			persistent: make([]kv, 0, mapSize),
			transient:  make([]kv, 0, mapSize),
			stale:      make([]kv, 0, mapSize),
		}
	}

	// insert new kvs from m to node
	for k, v := range m {
		if len(k) == 0 || len(v) == 0 {
			continue
		}
		switch {
		case strings.HasPrefix(k, PrefixTransientUpstream):
			if len(k) > lenPTU { // do not move this condition to the case statement to prevent a PTU matches PT
				if inherit {
					stale[k[lenPTU:]] = v
				} else {
					nd.stale = append(nd.stale, kv{key: k[lenPTU:], val: v})
				}
			}
		case strings.HasPrefix(k, PrefixTransient):
			if len(k) > lenPT {
				if inherit {
					transient[k[lenPT:]] = v
				} else {
					nd.transient = append(nd.transient, kv{key: k[lenPT:], val: v})
				}
			}
		case strings.HasPrefix(k, PrefixPersistent):
			if len(k) > lenPP {
				if inherit {
					persistent[k[lenPP:]] = v
				} else {
					nd.persistent = append(nd.persistent, kv{key: k[lenPP:], val: v})
				}
			}
		}
	}

	if nd.size() == 0 { // return original ctx if no invalid key in map
		return ctx
	}
	// transfer map to list
	if inherit {
		nd.stale = stale.toList(nd.stale)
		nd.transient = transient.toList(nd.transient)
		nd.persistent = persistent.toList(nd.persistent)
		stale.recycle()
		transient.recycle()
		persistent.recycle()
	}
	ctx = withNode(ctx, nd)
	return ctx
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
	for _, kv := range slice {
		kvs[kv.key] = kv.val
	}
}

// mapToSlice converts a map to a kv slice. If the map is empty, the return value will be nil.
func mapToSlice(kvs kvstore) (slice []kv) {
	size := len(kvs)
	if size == 0 {
		return
	}
	slice = make([]kv, 0, size)
	for k, v := range kvs {
		slice = append(slice, kv{key: k, val: v})
	}
	return
}
