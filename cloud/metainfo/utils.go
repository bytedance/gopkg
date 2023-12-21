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
	if ctx == nil {
		return nil
	}

	if len(m) == 0 {
		return ctx
	}

	n := &node{}
	if x := getNode(ctx); x != nil {
		copyNode(n, x)
	}

	for k, v := range m {
		if len(k) == 0 || len(v) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(k, PrefixTransientUpstream):
			if len(k) > lenPTU { // do not move this condition to the case statement to prevent a PTU matches PT
				if idx, ok := search(n.stale, k[lenPTU:]); ok {
					if n.stale[idx].val != v {
						n.stale[idx].val = v
					}
				} else {
					n.stale = append(n.stale, kv{k[lenPTU:], v})
				}
			}
		case strings.HasPrefix(k, PrefixTransient):
			if len(k) > lenPT {
				if idx, ok := search(n.transient, k[lenPT:]); ok {
					if n.transient[idx].val != v {
						n.transient[idx].val = v
					}
				} else {
					n.transient = append(n.transient, kv{k[lenPT:], v})
				}
			}
		case strings.HasPrefix(k, PrefixPersistent):
			if len(k) > lenPP {
				if idx, ok := search(n.persistent, k[lenPP:]); ok {
					if n.persistent[idx].val != v {
						n.persistent[idx].val = v
					}
				} else {
					n.persistent = append(n.persistent, kv{k[lenPP:], v})
				}
			}
		}
	}

	if n.size() == 0 {
		//TODO: remove this?
		return ctx
	}

	return withNode(ctx, n)
}

// SaveMetaInfoToMap set key-value pairs from ctx to m while filtering out transient-upstream data.
func SaveMetaInfoToMap(ctx context.Context, m map[string]string) {
	if ctx == nil || m == nil {
		return
	}
	ctx = TransferForward(ctx)
	for k, v := range GetAllValues(ctx) {
		m[PrefixTransient+k] = v
	}
	for k, v := range GetAllPersistentValues(ctx) {
		m[PrefixPersistent+k] = v
	}
}

// sliceToMap converts a kv slice to map. If the slice is empty, an empty map will be returned instead of nil.
func sliceToMap(slice []kv) (m map[string]string) {
	if size := len(slice); size == 0 {
		m = make(map[string]string)
	} else {
		m = make(map[string]string, size)
	}
	for _, kv := range slice {
		m[kv.key] = kv.val
	}
	return
}

// mapToSlice converts a map to a kv slice. If the map is empty, the return value will be nil.
func mapToSlice(kvs map[string]string) (slice []kv) {
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
