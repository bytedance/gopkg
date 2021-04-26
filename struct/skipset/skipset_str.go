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

package skipset

import (
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/bytedance/gopkg/internal/wyhash"
)

func hash(s string) uint64 {
	return wyhash.Sum64String(s)
}

// StringSet represents a set based on skip list.
// It based on Uint64Set.
type StringSet struct {
	header *stringNode
	tail   *stringNode
	length int64
}

type stringNode struct {
	value string
	score uint64
	next  []*stringNode
	mu    sync.Mutex
	flags bitflag
}

func newStringNode(value string, level int) *stringNode {
	return &stringNode{
		value: value,
		score: hash(value),
		next:  make([]*stringNode, level),
	}
}

// return n.next[i](atomic)
func (n *stringNode) loadNext(i int) *stringNode {
	return (*stringNode)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&n.next[i]))))
}

// n.next[i] = node(atomic)
func (n *stringNode) storeNext(i int, node *stringNode) {
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&n.next[i])), unsafe.Pointer(node))
}

// Return 1 if n is bigger, 0 if equal, else -1.
func (n *stringNode) cmp(score uint64, value string) int {
	if n.score > score {
		return 1
	} else if n.score == score {
		return cmpstring(n.value, value)
	}
	return -1
}

// NewString return an empty string skip set.
func NewString() *StringSet {
	h, t := newStringNode("", maxLevel), newStringNode("", maxLevel)
	for i := 0; i < maxLevel; i++ {
		h.next[i] = t
	}
	h.flags.SetTrue(fullyLinked)
	t.flags.SetTrue(fullyLinked)
	return &StringSet{
		header: h,
		tail:   t,
	}
}

// findNodeRemove takes a score and two maximal-height arrays then searches exactly as in a sequential skip-list.
// The returned preds and succs always satisfy preds[i] > score >= succs[i].
func (s *StringSet) findNodeRemove(value string, preds *[maxLevel]*stringNode, succs *[maxLevel]*stringNode) int {
	// lFound represents the index of the first layer at which it found a node.
	score := hash(value)
	lFound, x := -1, s.header
	for i := maxLevel - 1; i >= 0; i-- {
		succ := x.loadNext(i)
		for succ != s.tail && succ.cmp(score, value) < 0 {
			x = succ
			succ = x.loadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the score already in the skip list.
		if lFound == -1 && succ != s.tail && succ.cmp(score, value) == 0 {
			lFound = i
		}
	}
	return lFound
}

// findNodeAdd takes a score and two maximal-height arrays then searches exactly as in a sequential skip-set.
// The returned preds and succs always satisfy preds[i] > score >= succs[i].
func (s *StringSet) findNodeAdd(value string, preds *[maxLevel]*stringNode, succs *[maxLevel]*stringNode) int {
	// lFound represents the index of the first layer at which it found a node.
	score := hash(value)
	x := s.header
	for i := maxLevel - 1; i >= 0; i-- {
		succ := x.loadNext(i)
		for succ != s.tail && succ.cmp(score, value) < 0 {
			x = succ
			succ = x.loadNext(i)
		}
		preds[i] = x
		succs[i] = succ

		// Check if the score already in the skip list.
		if succ != s.tail && succ.cmp(score, value) == 0 {
			return i
		}
	}
	return -1
}

func unlockString(preds [maxLevel]*stringNode, highestLevel int) {
	var prevPred *stringNode
	for i := highestLevel; i >= 0; i-- {
		if preds[i] != prevPred { // the node could be unlocked by previous loop
			preds[i].mu.Unlock()
			prevPred = preds[i]
		}
	}
}

// Add add the score into skip set, return true if this process insert the score into skip set,
// return false if this process can't insert this score, because another process has insert the same score.
//
// If the score is in the skip set but not fully linked, this process will wait until it is.
func (s *StringSet) Add(value string) bool {
	level := randomLevel()
	var preds, succs [maxLevel]*stringNode
	for {
		lFound := s.findNodeAdd(value, &preds, &succs)
		if lFound != -1 { // indicating the score is already in the skip-list
			nodeFound := succs[lFound]
			if !nodeFound.flags.Get(marked) {
				for !nodeFound.flags.Get(fullyLinked) {
					// The node is not yet fully linked, just waits until it is.
				}
				return false
			}
			// If the node is marked, represents some other thread is in the process of deleting this node,
			// we need to add this node in next loop.
			continue
		}

		// Add this node into skip list.
		var (
			highestLocked        = -1 // the highest level being locked by this process
			valid                = true
			pred, succ, prevPred *stringNode
		)
		for layer := 0; valid && layer < level; layer++ {
			pred = preds[layer]   // target node's previous node
			succ = succs[layer]   // target node's next node
			if pred != prevPred { // the node in this layer could be locked by previous loop
				pred.mu.Lock()
				highestLocked = layer
				prevPred = pred
			}
			// valid check if there is another node has inserted into the skip list in this layer during this process.
			// It is valid if:
			// 1. The previous node and next node both are not marked.
			// 2. The previous node's next node is succ in this layer.
			valid = !pred.flags.Get(marked) && !succ.flags.Get(marked) && pred.next[layer] == succ
		}
		if !valid {
			unlockString(preds, highestLocked)
			continue
		}

		nn := newStringNode(value, level)
		for layer := 0; layer < level; layer++ {
			nn.next[layer] = succs[layer]
			preds[layer].storeNext(layer, nn)
		}
		nn.flags.SetTrue(fullyLinked)
		unlockString(preds, highestLocked)
		atomic.AddInt64(&s.length, 1)
		return true
	}
}

// Contains check if the score is in the skip set.
func (s *StringSet) Contains(value string) bool {
	score := hash(value)
	x := s.header
	for i := maxLevel - 1; i >= 0; i-- {
		nex := x.loadNext(i)
		for nex != s.tail && nex.cmp(score, value) < 0 {
			x = nex
			nex = x.loadNext(i)
		}

		// Check if the score already in the skip list.
		if nex != s.tail && nex.cmp(score, value) == 0 {
			return nex.flags.MGet(fullyLinked|marked, fullyLinked)
		}
	}
	return false
}

// Remove a node from the skip set.
func (s *StringSet) Remove(value string) bool {
	var (
		nodeToRemove *stringNode
		isMarked     bool // represents if this operation mark the node
		topLayer     = -1
		preds, succs [maxLevel]*stringNode
	)
	for {
		lFound := s.findNodeRemove(value, &preds, &succs)
		if isMarked || // this process mark this node or we can find this node in the skip list
			lFound != -1 && succs[lFound].flags.MGet(fullyLinked|marked, fullyLinked) && (len(succs[lFound].next)-1) == lFound {
			if !isMarked { // we don't mark this node for now
				nodeToRemove = succs[lFound]
				topLayer = lFound
				nodeToRemove.mu.Lock()
				if nodeToRemove.flags.Get(marked) {
					// The node is marked by another process,
					// the physical deletion will be accomplished by another process.
					nodeToRemove.mu.Unlock()
					return false
				}
				nodeToRemove.flags.SetTrue(marked)
				isMarked = true
			}
			// Accomplish the physical deletion.
			var (
				highestLocked        = -1 // the highest level being locked by this process
				valid                = true
				pred, succ, prevPred *stringNode
			)
			for layer := 0; valid && (layer <= topLayer); layer++ {
				pred, succ = preds[layer], succs[layer]
				if pred != prevPred { // the node in this layer could be locked by previous loop
					pred.mu.Lock()
					highestLocked = layer
					prevPred = pred
				}
				// valid check if there is another node has inserted into the skip list in this layer
				// during this process, or the previous is removed by another process.
				// It is valid if:
				// 1. the previous node exists.
				// 2. no another node has inserted into the skip list in this layer.
				valid = !pred.flags.Get(marked) && pred.next[layer] == succ
			}
			if !valid {
				unlockString(preds, highestLocked)
				continue
			}
			for i := topLayer; i >= 0; i-- {
				// Now we own the `nodeToRemove`, no other goroutine will modify it.
				// So we don't need `nodeToRemove.loadNext`
				preds[i].storeNext(i, nodeToRemove.next[i])
			}
			nodeToRemove.mu.Unlock()
			unlockString(preds, highestLocked)
			atomic.AddInt64(&s.length, -1)
			return true
		}
		return false
	}
}

// Range calls f sequentially for each val present in the skip set.
// If f returns false, range stops the iteration.
func (s *StringSet) Range(f func(value string) bool) {
	x := s.header.loadNext(0)
	for x != s.tail {
		if !x.flags.MGet(fullyLinked|marked, fullyLinked) {
			x = x.loadNext(0)
			continue
		}
		if !f(x.value) {
			break
		}
		x = x.loadNext(0)
	}
}

// Len return the length of this skip set.
// Keep in sync with types_gen.go:lengthFunction
// Special case for code generation, Must in the tail of skipset.go.
func (s *StringSet) Len() int {
	return int(atomic.LoadInt64(&s.length))
}
