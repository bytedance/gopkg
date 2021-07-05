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

package msq

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

var msqv1pool *sync.Pool = &sync.Pool{New: func() interface{} { return new(msqv1node) }}

type MSQueue struct {
	head unsafe.Pointer // *msqv1node
	tail unsafe.Pointer // *msqv1node
}

type msqv1node struct {
	value uint64
	next  unsafe.Pointer // *msqv1node
}

func New() *MSQueue {
	node := unsafe.Pointer(new(msqv1node))
	return &MSQueue{head: node, tail: node}
}

func loadMSQPointer(p *unsafe.Pointer) *msqv1node {
	return (*msqv1node)(atomic.LoadPointer(p))
}

func (q *MSQueue) Enqueue(value uint64) bool {
	node := &msqv1node{value: value}
	for {
		tail := atomic.LoadPointer(&q.tail)
		tailstruct := (*msqv1node)(tail)
		next := atomic.LoadPointer(&tailstruct.next)
		if tail == atomic.LoadPointer(&q.tail) {
			if next == nil {
				// tail.next is empty, inset new node
				if atomic.CompareAndSwapPointer(&tailstruct.next, next, unsafe.Pointer(node)) {
					atomic.CompareAndSwapPointer(&q.tail, tail, unsafe.Pointer(node))
					break
				}
			} else {
				atomic.CompareAndSwapPointer(&q.tail, tail, next)
			}
		}
	}
	return true
}

func (q *MSQueue) Dequeue() (value uint64, ok bool) {
	for {
		head := atomic.LoadPointer(&q.head)
		tail := atomic.LoadPointer(&q.tail)
		headstruct := (*msqv1node)(head)
		next := atomic.LoadPointer(&headstruct.next)
		if head == atomic.LoadPointer(&q.head) {
			if head == tail {
				if next == nil {
					return 0, false
				}
				atomic.CompareAndSwapPointer(&q.tail, tail, next)
			} else {
				value = ((*msqv1node)(next)).value
				if atomic.CompareAndSwapPointer(&q.head, head, next) {
					return value, true
				}
			}
		}
	}
}
