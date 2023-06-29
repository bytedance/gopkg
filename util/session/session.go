/**
 * Copyright 2023 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package session

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Session represents a local storage for one session
type Session interface {
    // IsValid tells if the session is valid at present
    IsValid() bool 

    // Get returns value for specific key
    Get(key interface{}) interface{}
    
    // WithValue sets value for specific key，and return newly effective session
    WithValue(key interface{}, val interface{}) Session
}

// SessionCtx implements Session with context, 
// which means children session WON'T affect parent and sibling sessions
type SessionCtx struct {
	enabled *atomic.Value
	storage context.Context
}

// NewSessionCtx creates and enables a SessionCtx
func NewSessionCtx(ctx context.Context) *SessionCtx {
	var enabled atomic.Value
	enabled.Store(true)
	return &SessionCtx{
		enabled: &enabled,
		storage: ctx,
	}
}

// NewSessionCtx creates and enables a SessionCtx,
// and disable the session after timeout
func NewSessionCtxWithTimeout(ctx context.Context, timeout time.Duration) *SessionCtx {
	ret := NewSessionCtx(ctx)
	go func() {
		<- time.NewTimer(timeout).C
		ret.Disable()
	}()
	return ret
}

// Disable ends the session
func (self SessionCtx) Disable() {
	self.enabled.Store(false)
}

// IsValid tells if the session is valid at present
func (self *SessionCtx) IsValid() bool {
	if self == nil {
		return false
	}
	return self.enabled.Load().(bool)
}

// Get value for specific key
func (self *SessionCtx) Get(key interface{}) interface{} {
	if self == nil {
		return nil
	}
	return self.storage.Value(key)
}

// Set value for specific key，and return newly effective session
func (self SessionCtx) WithValue(key interface{}, val interface{}) Session {
	ctx := context.WithValue(self.storage, key, val)
	return &SessionCtx{
		enabled: self.enabled,
		storage: ctx,
	}
}

// NewSessionMap implements Session with map, 
// which means children session WILL affect parent session and sibling sessions
type SessionMap struct {
	enabled *atomic.Value
	storage map[interface{}]interface{}
	lock    sync.RWMutex
}

// NewSessionMap creates and enables a SessionMap
func NewSessionMap(m map[interface{}]interface{}) *SessionMap {
	var enabled atomic.Value
	enabled.Store(true)
	return &SessionMap{
		enabled: &enabled,
		storage: m,
	}
}

// NewSessionCtx creates and enables a SessionCtx,
// and disable the session after timeout
func NewSessionMapWithTimeout(m map[interface{}]interface{}, timeout time.Duration) *SessionMap {
	ret := NewSessionMap(m)
	go func() {
		<- time.NewTimer(timeout).C
		ret.Disable()
	}()
	return ret
}

// IsValid tells if the session is valid at present
func (self *SessionMap) IsValid() bool {
	if self == nil {
		return false
	}
	return self.enabled.Load().(bool)
}

// Disable ends the session
func (self *SessionMap) Disable() {
	self.enabled.Store(false)
}

// Get value for specific key
func (self *SessionMap) Get(key interface{}) interface{} {
	if self == nil {
		return nil
	}
	self.lock.RLock()
	val := self.storage[key]
	self.lock.RUnlock()
	return val
}

// Set value for specific key，and return itself
func (self *SessionMap) WithValue(key interface{}, val interface{}) Session {
	self.lock.Lock()
	self.storage[key] = val
	self.lock.Unlock()
	return self
}
