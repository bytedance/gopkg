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
func SetMetaInfoFromMap(ctx context.Context, m map[string]string) context.Context {
	if ctx == nil {
		return nil
	}

	var n node
	for k, v := range m {
		if len(k) == 0 || len(v) == 0 {
			continue
		}

		switch {
		case strings.HasPrefix(k, PrefixTransientUpstream):
			if len(k) > lenPTU { // do not move this condition to the case statement to prevent a PTU matches PT
				n.stale = append(n.stale, kv{key: k[lenPTU:], val: v})
			}
		case strings.HasPrefix(k, PrefixTransient):
			if len(k) > lenPT {
				n.transient = append(n.transient, kv{key: k[lenPT:], val: v})
			}
		case strings.HasPrefix(k, PrefixPersistent):
			if len(k) > lenPP {
				n.persistent = append(n.persistent, kv{key: k[lenPP:], val: v})
			}
		}
	}

	if n.size() == 0 {
		// TODO: remove this?
		return ctx
	}
	return withNode(ctx, &n)
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
