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
	"fmt"
	"time"
)

const (
	// DefaultShardNum set the sharding number of id->sesssion map for default SessionManager
	DefaultShardNum = 100
	
	// DefaultGCInterval set the GC interval for default SessionManager
	DefaultGCInterval = time.Hour

	// DefaultEnableTransparentTransmitAsync enables TransparentTransmitAsync for default SessionManager
	DefaultEnableTransparentTransmitAsync = false
)

var (
	defaultManagerObj *SessionManager
)

func init() {
	obj := NewSessionManager(ManagerOptions{
		EnableTransparentTransmitAsync: DefaultEnableTransparentTransmitAsync,
		ShardNumber: DefaultShardNum,
		GCInterval: DefaultGCInterval,
	})
	defaultManagerObj = &obj
}

// SetDefaultManager updates default SessionManager to m
func SetDefaultManager(m SessionManager) {
	if defaultManagerObj != nil {
		defaultManagerObj.Close()
	}
	defaultManagerObj = &m
}

// GetDefaultManager returns a copy of default SessionManager
//   warning: use it only for state check
func GetDefaultManager() SessionManager {
	return *defaultManagerObj
}

// CurSession gets the session for current goroutine
func CurSession() (Session, bool) {
	s, ok := defaultManagerObj.GetSession(SessionID(goID()))
	return s, ok
}

// BindSession binds the session with current goroutine
func BindSession(s Session) {
	defaultManagerObj.BindSession(SessionID(goID()), s)
}

// UnbindSession unbind a session (if any) with current goroutine
//
// Notice: If you want to end the session, 
// please call `Disable()` (or whatever make the session invalid)
// on your session's implementation
func UnbindSession() {
	defaultManagerObj.UnbindSession(SessionID(goID()))
}

// Go calls f asynchronously and pass caller's session to the new goroutine
func Go(f func()) {
	s, ok := CurSession()
	if !ok {
		GoSession(nil, f)
	} else {
		GoSession(s, f)
	}
}

// SessionGo calls f asynchronously and pass s session to the new goroutine
func GoSession(s Session, f func()) {
	go func(){
		defer func() {
			if v := recover(); v != nil {
				println(fmt.Sprintf("GoSession recover: %v", v))
			}
			UnbindSession()
		}()
		if s != nil {
			BindSession(s)
		}
		f()
	}()
}
