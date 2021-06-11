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
	if q.tail.next == nil {
		q.tail.next = &linkedqueueNode{value: value}
		q.tail = q.tail.next
	}
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
