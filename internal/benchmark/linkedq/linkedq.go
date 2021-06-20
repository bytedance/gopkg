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

package linkedq

import "sync"

type LinkedQueue struct {
	head *linkedqueueNode
	tail *linkedqueueNode
	mu   sync.Mutex
}

type linkedqueueNode struct {
	value uint64
	next  *linkedqueueNode
}

func New() *LinkedQueue {
	node := new(linkedqueueNode)
	return &LinkedQueue{head: node, tail: node}
}

func (q *LinkedQueue) Enqueue(value uint64) bool {
	q.mu.Lock()
	q.tail.next = &linkedqueueNode{value: value}
	q.tail = q.tail.next
	q.mu.Unlock()
	return true
}

func (q *LinkedQueue) Dequeue() (uint64, bool) {
	q.mu.Lock()
	if q.head.next == nil {
		q.mu.Unlock()
		return 0, false
	} else {
		value := q.head.next.value
		q.head = q.head.next
		q.mu.Unlock()
		return value, true
	}
}
