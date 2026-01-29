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

package gopoolsession

import (
	"context"
	"sync"
	"testing"

	"github.com/cloudwego/localsession"
	"github.com/stretchr/testify/assert"
)

// MockSession is a mock implementation of localsession.Session for testing
type MockSession struct {
	values map[interface{}]interface{}
}

// IsValid returns whether the session is valid
func (s *MockSession) IsValid() bool {
	return true
}

// Get returns the value associated with the key
func (s *MockSession) Get(key interface{}) interface{} {
	if s.values == nil {
		return nil
	}
	return s.values[key]
}

// WithValue returns a new session with the key-value pair added
func (s *MockSession) WithValue(key interface{}, val interface{}) localsession.Session {
	newSession := &MockSession{
		values: make(map[interface{}]interface{}),
	}
	for k, v := range s.values {
		newSession.values[k] = v
	}
	newSession.values[key] = val
	return newSession
}

// Value returns the value associated with the key (legacy method for compatibility)
func (s *MockSession) Value(key string) interface{} {
	return s.Get(key)
}

// SetValue sets the value associated with the key (legacy method for compatibility)
func (s *MockSession) SetValue(key string, value interface{}) {
	if s.values == nil {
		s.values = make(map[interface{}]interface{})
	}
	s.values[key] = value
}

// Reset resets the session (legacy method for compatibility)
func (s *MockSession) Reset() {
	s.values = nil
}

func TestGoWithSession(t *testing.T) {
	// Create a mock session and set it in the current goroutine
	localsession.InitDefaultManager(localsession.DefaultManagerOptions())
	ctx := context.Background()
	ctx = context.WithValue(ctx, "test-key", "test-value")
	mockSession := localsession.NewSessionCtx(ctx)
	localsession.BindSession(mockSession)
	defer localsession.UnbindSession()

	var wg sync.WaitGroup
	wg.Add(1)

	var taskSession localsession.Session
	var ok bool

	// Submit a task using gopoolsession.Go
	Go(func() {
		defer wg.Done()
		// Get the session in the task goroutine
		taskSession, ok = localsession.CurSession()
	})

	wg.Wait()

	// Verify that the session was correctly propagated
	assert.NotNil(t, taskSession)
	assert.Equal(t, "test-value", taskSession.Get("test-key"))

	// Verify that the session was unbound after task completion
	currentSession, _ := localsession.CurSession()
	assert.Equal(t, mockSession, currentSession) // The original session should still be bound in the main goroutine
	assert.True(t, ok)
}

func TestCtxGoWithSession(t *testing.T) {
	// Create a mock session and set it in the current goroutine
	localsession.InitDefaultManager(localsession.DefaultManagerOptions())
	ctx := context.Background()
	ctx = context.WithValue(ctx, "ctx-key", "ctx-value")
	mockSession := localsession.NewSessionCtx(ctx)
	localsession.BindSession(mockSession)
	defer localsession.UnbindSession()

	var wg sync.WaitGroup
	wg.Add(1)

	var taskSession localsession.Session
	var ok bool

	// Submit a task using gopoolsession.CtxGo with context
	ctx = context.Background()
	CtxGo(ctx, func() {
		defer wg.Done()
		// Get the session in the task goroutine
		taskSession, ok = localsession.CurSession()
	})

	wg.Wait()

	// Verify that the session was correctly propagated
	assert.NotNil(t, taskSession)
	assert.True(t, ok)
	assert.Equal(t, "ctx-value", taskSession.Get("ctx-key"))
}

func TestGoWithoutSession(t *testing.T) {
	// Ensure no session is bound in the current goroutine
	localsession.InitDefaultManager(localsession.DefaultManagerOptions())
	localsession.UnbindSession()

	var wg sync.WaitGroup
	wg.Add(1)

	var taskSession localsession.Session
	var ok bool

	// Submit a task using gopoolsession.Go
	Go(func() {
		defer wg.Done()
		// Get the session in the task goroutine
		taskSession, ok = localsession.CurSession()
	})

	wg.Wait()

	// Verify that no session was propagated
	assert.Nil(t, taskSession)
	assert.False(t, ok)
}

func TestSessionUnbindingAfterPanic(t *testing.T) {
	localsession.InitDefaultManager(localsession.DefaultManagerOptions())
	ctx := context.Background()
	ctx = context.WithValue(ctx, "panic-key", "panic-value")
	mockSession := localsession.NewSessionCtx(ctx)
	localsession.BindSession(mockSession)
	defer localsession.UnbindSession()

	var wg sync.WaitGroup
	wg.Add(1)

	// Submit a task that panics
	Go(func() {
		defer func() {
			if r := recover(); r != nil {
				// Recover from panic
			}
			wg.Done()
		}()
		// Get the session to verify it was bound
		sess, _ := localsession.CurSession()
		assert.NotNil(t, sess)
		// Panic intentionally
		panic("test panic")
	})

	wg.Wait()

	// Verify that we can still get the session in the main goroutine
	mainSession, _ := localsession.CurSession()
	assert.Equal(t, mockSession, mainSession)
}

func TestMultipleTasksWithSession(t *testing.T) {
	// Create a mock session and set it in the current goroutine
	mockSession := &MockSession{
		values: map[interface{}]interface{}{"multi-key": "multi-value"},
	}
	localsession.BindSession(mockSession)
	defer localsession.UnbindSession()

	const taskCount = 100
	var wg sync.WaitGroup
	wg.Add(taskCount)

	var sessions []localsession.Session
	var mu sync.Mutex

	// Submit multiple tasks using gopoolsession.Go
	for i := 0; i < taskCount; i++ {
		Go(func() {
			defer wg.Done()
			// Get the session in the task goroutine
			taskSession, _ := localsession.CurSession()
			mu.Lock()
			sessions = append(sessions, taskSession)
			mu.Unlock()
		})
	}

	wg.Wait()

	// Verify that all tasks received the correct session
	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, sessions, taskCount)
	for _, sess := range sessions {
		assert.NotNil(t, sess)
		assert.Equal(t, "multi-value", sess.Get("multi-key"))
	}
}
