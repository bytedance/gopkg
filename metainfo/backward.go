// Copyright [2021] [ByteDance Inc.]
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
	"sync"
)

type bwCtxKeyType struct{}

var bwCtxKey bwCtxKeyType

// WithBackwardValues returns a new context that allows passing key-value pairs backward from any derived context.
func WithBackwardValues(ctx context.Context) context.Context {
	if _, ok := ctx.Value(bwCtxKey).(*sync.Map); ok {
		return ctx
	}
	ctx = context.WithValue(ctx, bwCtxKey, new(sync.Map))
	return ctx
}

// SetBackwardValue .
func SetBackwardValue(ctx context.Context, k, v string) (ok bool) {
	if kvs, ok := ctx.Value(bwCtxKey).(*sync.Map); ok {
		kvs.Store(k, v)
		return true
	}
	return false
}

// GetBackwardValue .
func GetBackwardValue(ctx context.Context, key string) (val string, ok bool) {
	if kvs, ok := ctx.Value(bwCtxKey).(*sync.Map); ok {
		if v, ok := kvs.Load(key); ok {
			return v.(string), true
		}
	}
	return "", false
}

// GetAllBackwardValues retrieves all key-value pairs set by SetBackwardValue from the given context.
// If the context is not created by WithBackwardValues, the result will be nil.
func GetAllBackwardValues(ctx context.Context) map[string]string {
	if kvs, ok := ctx.Value(bwCtxKey).(*sync.Map); ok {
		m := make(map[string]string)
		kvs.Range(func(key, val interface{}) bool {
			m[key.(string)] = val.(string)
			return true
		})
		return m
	}
	return nil
}
