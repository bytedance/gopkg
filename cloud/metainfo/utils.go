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
	return getKV(ctx) != nil
}

// SetMetaInfoFromMap retrieves metainfo key-value pairs from the given map and sets then into the context.
// Only those keys with prefixes defined in this module would be used.
func SetMetaInfoFromMap(ctx context.Context, m map[string]string) context.Context {
	if ctx == nil {
		return nil
	}
	for k, v := range m {
		if t, nk := determineKeyType(k, v); t != invalidType {
			ctx = addKV(ctx, t, nk, v)
		}
	}
	return ctx
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

const (
	lenPTU = len(PrefixTransientUpstream)
	lenPT  = len(PrefixTransient)
	lenPP  = len(PrefixPersistent)
)

// determineKeyType tests whether the given key-value pair is a valid metainfo and returns its info type with a new appropriate key.
func determineKeyType(k, v string) (infoType infoType, newKey string) {
	if len(k) == 0 || len(v) == 0 {
		return invalidType, k
	}

	switch {
	case strings.HasPrefix(k, PrefixTransientUpstream):
		if len(k) > lenPTU {
			return transientUpstreamType, k[lenPTU:]
		}
	case strings.HasPrefix(k, PrefixTransient):
		if len(k) > lenPT {
			return transientType, k[lenPT:]
		}
	case strings.HasPrefix(k, PrefixPersistent):
		if len(k) > lenPP {
			return persistentType, k[lenPP:]
		}
	}
	return invalidType, k
}
