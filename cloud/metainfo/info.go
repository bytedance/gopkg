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
)

// The prefix listed below may be used to tag the types of values when there is no context to carry them.
const (
	PrefixPersistent         = "RPC_PERSIST_"
	PrefixTransient          = "RPC_TRANSIT_"
	PrefixTransientUpstream  = "RPC_TRANSIT_UPSTREAM_"
	PrefixBackward           = "RPC_BACKWARD_"
	PrefixBackwardDownstream = "RPC_BACKWARD_DOWNSTREAM_"

	lenPTU = len(PrefixTransientUpstream)
	lenPT  = len(PrefixTransient)
	lenPP  = len(PrefixPersistent)
	lenB   = len(PrefixBackward)
	lenBD  = len(PrefixBackwardDownstream)
)

// **Using empty string as key or value is not support.**

// TransferForward converts transient values to transient-upstream values and filters out original transient-upstream values.
// It should be used before the context is passing from server to client.
func TransferForward(ctx context.Context) context.Context {
	if n := getNode(ctx); n != nil {
		return withNode(ctx, n.transferForward())
	}
	return ctx
}

// GetValue retrieves the value set into the context by the given key.
func GetValue(ctx context.Context, k string) (v string, ok bool) {
	if n := getNode(ctx); n != nil {
		if idx, ok := search(n.transient, k); ok {
			return n.transient[idx].val, true
		}
		if idx, ok := search(n.stale, k); ok {
			return n.stale[idx].val, true
		}
	}
	return
}

// GetAllValues retrieves all transient values
func GetAllValues(ctx context.Context) (m map[string]string) {
	if n := getNode(ctx); n != nil {
		if cnt := len(n.stale) + len(n.transient); cnt > 0 {
			m = make(map[string]string, cnt)
			for _, kv := range n.stale {
				m[kv.key] = kv.val
			}
			for _, kv := range n.transient {
				m[kv.key] = kv.val
			}
		}
	}
	return
}

// GetValueToMap retrieves the value set into the context by the given key and set the value to the input map.
// Only use this function when you want to get a small set of values instead of GetAllValues.
// The logic of getting value follows GetAllValues, transient value has higher priority if key is same.
func GetValueToMap(ctx context.Context, m map[string]string, keys ...string) {
	if m == nil || len(keys) == 0 {
		return
	}
	n := getNode(ctx)
	if n == nil {
		return
	}
	for _, k := range keys {
		if idx, ok := search(n.transient, k); ok {
			m[k] = n.transient[idx].val
			continue
		}
		if idx, ok := search(n.stale, k); ok {
			m[k] = n.stale[idx].val
		}
	}
}

// RangeValues calls f sequentially for each transient kv.
// If f returns false, range stops the iteration.
func RangeValues(ctx context.Context, f func(k, v string) bool) {
	n := getNode(ctx)
	if n == nil {
		return
	}

	if cnt := len(n.stale) + len(n.transient); cnt == 0 {
		return
	}

	for _, kv := range n.stale {
		if !f(kv.key, kv.val) {
			return
		}
	}

	for _, kv := range n.transient {
		if !f(kv.key, kv.val) {
			return
		}
	}
}

// WithValue sets the value into the context by the given key.
// This value will be propagated to the next service/endpoint through an RPC call.
//
// Notice that it will not propagate any further beyond the next service/endpoint,
// Use WithPersistentValue if you want to pass a key/value pair all the way.
func WithValue(ctx context.Context, k, v string) context.Context {
	if len(k) == 0 || len(v) == 0 {
		return ctx
	}
	if n := getNode(ctx); n != nil {
		if m := n.addTransient(k, v); m != n {
			return withNode(ctx, m)
		}
	} else {
		return withNode(ctx, &node{
			transient: []kv{{key: k, val: v}},
		})
	}
	return ctx
}

// DelValue deletes a key/value from the current context.
// Since empty string value is not valid, we could just set the value to be empty.
func DelValue(ctx context.Context, k string) context.Context {
	if len(k) == 0 {
		return ctx
	}
	if n := getNode(ctx); n != nil {
		if m := n.delTransient(k); m != n {
			return withNode(ctx, m)
		}
	}
	return ctx
}

// GetPersistentValue retrieves the persistent value set into the context by the given key.
func GetPersistentValue(ctx context.Context, k string) (v string, ok bool) {
	if n := getNode(ctx); n != nil {
		if idx, ok := search(n.persistent, k); ok {
			return n.persistent[idx].val, true
		}
	}
	return
}

// GetAllPersistentValues retrieves all persistent values.
func GetAllPersistentValues(ctx context.Context) (m map[string]string) {
	if n := getNode(ctx); n != nil {
		if cnt := len(n.persistent); cnt > 0 {
			m = make(map[string]string, cnt)
			for _, kv := range n.persistent {
				m[kv.key] = kv.val
			}
		}
	}
	return
}

// RangePersistentValues calls f sequentially for each persistent kv.
// If f returns false, range stops the iteration.
func RangePersistentValues(ctx context.Context, f func(k, v string) bool) {
	n := getNode(ctx)
	if n == nil {
		return
	}

	for _, kv := range n.persistent {
		if !f(kv.key, kv.val) {
			break
		}
	}
}

// WithPersistentValue sets the value into the context by the given key.
// This value will be propagated to the services along the RPC call chain.
func WithPersistentValue(ctx context.Context, k, v string) context.Context {
	if len(k) == 0 || len(v) == 0 {
		return ctx
	}
	if n := getNode(ctx); n != nil {
		if m := n.addPersistent(k, v); m != n {
			return withNode(ctx, m)
		}
	} else {
		return withNode(ctx, &node{
			persistent: []kv{{key: k, val: v}},
		})
	}
	return ctx
}

// DelPersistentValue deletes a persistent key/value from the current context.
// Since empty string value is not valid, we could just set the value to be empty.
func DelPersistentValue(ctx context.Context, k string) context.Context {
	if len(k) == 0 {
		return ctx
	}
	if n := getNode(ctx); n != nil {
		if m := n.delPersistent(k); m != n {
			return withNode(ctx, m)
		}
	}
	return ctx
}

func getKey(kvs []string, i int) string {
	return kvs[i*2]
}

func getValue(kvs []string, i int) string {
	return kvs[i*2+1]
}

// CountPersistentValues counts the length of persisten KV pairs
func CountPersistentValues(ctx context.Context) int {
	if n := getNode(ctx); n == nil {
		return 0
	} else {
		return len(n.persistent)
	}
}

// CountValues counts the length of transient KV pairs
func CountValues(ctx context.Context) int {
	if n := getNode(ctx); n == nil {
		return 0
	} else {
		return len(n.stale) + len(n.transient)
	}
}

// WithPersistentValues sets the values into the context by the given keys.
// This value will be propagated to the services along the RPC call chain.
func WithPersistentValues(ctx context.Context, kvs ...string) context.Context {
	if len(kvs)%2 != 0 {
		panic("len(kvs) must be even")
	}

	kvLen := len(kvs) / 2

	if ctx == nil || len(kvs) == 0 {
		return ctx
	}

	var n *node
	if m := getNode(ctx); m != nil {
		nn := *m
		n = &nn
		n.persistent = make([]kv, len(m.persistent), len(m.persistent)+kvLen)
		copy(n.persistent, m.persistent)
	} else {
		n = &node{
			persistent: make([]kv, 0, kvLen),
		}
	}

	for i := 0; i < kvLen; i++ {
		key := getKey(kvs, i)
		val := getValue(kvs, i)

		if len(key) == 0 || len(val) == 0 {
			continue
		}

		if idx, ok := search(n.persistent, key); ok {
			if n.persistent[idx].val != val {
				n.persistent[idx].val = val
			}
		} else {
			n.persistent = append(n.persistent, kv{key: key, val: val})
		}
	}

	return withNode(ctx, n)
}

// WithValue sets the values into the context by the given keys.
// This value will be propagated to the next service/endpoint through an RPC call.
//
// Notice that it will not propagate any further beyond the next service/endpoint,
// Use WithPersistentValues if you want to pass key/value pairs all the way.
func WithValues(ctx context.Context, kvs ...string) context.Context {
	if len(kvs)%2 != 0 {
		panic("len(kvs) must be even")
	}

	kvLen := len(kvs) / 2

	if ctx == nil || len(kvs) == 0 {
		return ctx
	}

	var n *node
	if m := getNode(ctx); m != nil {
		nn := *m
		n = &nn
		n.transient = make([]kv, len(m.transient), len(m.transient)+kvLen)
		copy(n.transient, m.transient)
	} else {
		n = &node{
			transient: make([]kv, 0, kvLen),
		}
	}

	for i := 0; i < kvLen; i++ {
		key := getKey(kvs, i)
		val := getValue(kvs, i)

		if len(key) == 0 || len(val) == 0 {
			continue
		}

		if res, ok := remove(n.stale, key); ok {
			n.stale = res
			n.transient = append(n.transient, kv{key: key, val: val})
			continue
		}

		if idx, ok := search(n.transient, key); ok {
			if n.transient[idx].val != val {
				n.transient[idx].val = val
			}
		} else {
			n.transient = append(n.transient, kv{key: key, val: val})
		}
	}

	return withNode(ctx, n)
}
