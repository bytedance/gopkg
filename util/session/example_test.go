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
)

func ASSERT(v bool) {
	if !v {
		panic("not true!")
	}
}

func GetCurSession() Session {
	s, ok := CurSession()
	if !ok {
		panic("can't get current seession!")
	}
	return s
}

func ExampleSessionCtx() {
	var ctx = context.Background()
	var key, v = "a", "b"
	var key2, v2 = "c", "d"
	var sig = make(chan struct{})
	var sig2 = make(chan struct{})

	// initialize new session with context
	var session = NewSessionCtx(ctx)// implementation...

	// set specific key-value and update session
	start := session.WithValue(key, v)

	// set current session
	BindSession(start)

	// pass to new goroutine...
	Go(func(){
		// read specific key under current session
		val := GetCurSession().Get(key) // val exists
		ASSERT(val == v)
		// doSomething....
		
		// set specific key-value under current session
		// NOTICE: current session won't change here
		next := GetCurSession().WithValue(key2, v2)
		val2 := GetCurSession().Get(key2) // val2 == nil
		ASSERT(val2 == nil)
		
		// pass both parent session and new session to sub goroutine
		GoSession(next, func(){
			// read specific key under current session
			val := GetCurSession().Get(key) // val exists
			ASSERT(val == v)

			val2 := GetCurSession().Get(key2) // val2 exists
			ASSERT(val2 == v2)
			// doSomething....
			
			sig2 <- struct{}{}

			<- sig
			ASSERT(GetCurSession().IsValid() == false) // current session is invalid
			
			println("g2 done")
			sig2 <- struct{}{}
		})
		
		Go(func() {
			// read specific key under current session
			val := GetCurSession().Get(key) // val exists
			ASSERT(v == val)

			val2 := GetCurSession().Get(key2) // val2 == nil
			ASSERT(val2 == nil)
			// doSomething....

			sig2 <- struct{}{}

			<- sig
			ASSERT(GetCurSession().IsValid() == false) // current session is invalid

			println("g3 done")
			sig2 <- struct{}{}
		})
		
		BindSession(next)
		val2 = GetCurSession().Get(key2) // val2 exists
		ASSERT(v2 == val2)

		sig2 <- struct{}{}

		<- sig
		ASSERT(next.IsValid() == false) // next is invalid

		println("g1 done")
		sig2 <- struct{}{}
	})

	<- sig2
	<- sig2
	<- sig2

	val2 := GetCurSession().Get(key2) // val2 == nil
	ASSERT(val2 == nil)

	// initiatively ends the session，
	// then all the inherited session (including next) will be disabled
	session.Disable()
	close(sig)

	ASSERT(start.IsValid() == false) // start is invalid
	
	<- sig2
	<- sig2
	<- sig2
	println("g0 done")

	UnbindSession()
}