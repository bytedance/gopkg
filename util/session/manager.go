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
	"sync"
	"sync/atomic"
	"time"
)

// ManagerOptions for SessionManager
type ManagerOptions struct {
	// EnableTransparentTransmitAsync enables transparently transmit 
	// current session to children goroutines
	EnableTransparentTransmitAsync bool
	// ShardNumber is used to shard session id, it must be larger than zero 
	ShardNumber int
	// GCInterval decides the GC interval for SessionManager, 
	// it must be larger than 1s or zero means disable GC
	GCInterval time.Duration
}

type shard struct {
	lock sync.RWMutex
	m    map[SessionID]Session
}

// SessionManager maintain and manage sessions
type SessionManager struct {
	shards []*shard
	inGC   uint32
	opts   ManagerOptions 
}

var defaultShardCap = 10

func newShard() *shard {
	ret := new(shard)
	ret.m = make(map[SessionID]Session, defaultShardCap)
	return ret
}

// NewSessionManager creates a SessionManager with default containers
// If opts.GCInterval > 0, it will start scheduled GC() loop automatically 
func NewSessionManager(opts ManagerOptions) SessionManager {
	if opts.ShardNumber <= 0 {
		panic("ShardNumber must be larger than zero")
	}
	shards := make([]*shard, opts.ShardNumber)
	for i := range shards {
		shards[i] = newShard()
	}
	ret := SessionManager{
		shards: shards,
		opts  : opts,
	}

	if opts.GCInterval > 0 {
		ret.startGC()
	}
	return ret
}

// SessionID is the indentity of a session
type SessionID uint64

// Get gets specific session 
// or get inherited session if option EnableTransparentTransmitAsync is true
func (self SessionManager) GetSession(id SessionID) (Session, bool) {
	shard := self.shards[uint64(id)%uint64(self.opts.ShardNumber)]
	shard.lock.RLock()
	session, ok := shard.m[id]
	shard.lock.RUnlock()
	if ok {
		return session, ok
	}
	if !self.opts.EnableTransparentTransmitAsync {
		return nil, false
	}
	
	id, ok = getSessionID()
	if !ok {
		return nil, false
	}
	shard = self.shards[uint64(id)%uint64(self.opts.ShardNumber)]
	shard.lock.RLock()
	session, ok = shard.m[id]
	shard.lock.RUnlock()
	return session, ok
}

// BindSession binds the session with current goroutine
func (self SessionManager) BindSession(id SessionID, s Session) {
	shard := self.shards[uint64(id)%uint64(self.opts.ShardNumber)]
	shard.lock.Lock()
	shard.m[id] = s
	shard.lock.Unlock()

	if self.opts.EnableTransparentTransmitAsync {
		transmitSessionID(id)
	}
}

// UnbindSession clears current session
//
// Notice: If you want to end the session, 
// please call `Disable()` (or whatever make the session invalid)
// on your session's implementation
func (self SessionManager) UnbindSession(id SessionID) {
	shard := self.shards[uint64(id)%uint64(self.opts.ShardNumber)]
	
	shard.lock.RLock()
	_, ok := shard.m[id]
	shard.lock.RUnlock()
	if ok {
		shard.lock.Lock()
		delete(shard.m, id)
		shard.lock.Unlock()
	}
	
	if self.opts.EnableTransparentTransmitAsync {
		clearSessionID()
	}
}

// GC sweep invalid sessions and release unused memory
func (self SessionManager) GC() {
	if !atomic.CompareAndSwapUint32(&self.inGC, 0, 1) {
		return
	}

	for _, shard := range self.shards {
		shard.lock.RLock()
		m := make(map[SessionID]Session, len(shard.m))
		for id, s := range shard.m {
			// Warning: may panic here?
			if s.IsValid() {
				m[id] = s
			}
		}
		shard.m = m
		shard.lock.RUnlock()
	}

	atomic.StoreUint32(&self.inGC, 0)
}

// startGC start a scheduled goroutine to call GC() according to GCInterval
func (self SessionManager) startGC() {
	if self.opts.GCInterval < time.Second {
		panic("GCInterval must be larger than 1 second")
	}
	go func() {
		for range time.Tick(self.opts.GCInterval) {
			self.GC()
		}
	}()
}
