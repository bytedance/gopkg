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

// Package zset provides a concurrent-safety sorted set, can be used as a local
// replacement of Redis' zset (https://redis.com/ebook/part-2-core-concepts/chapter-3-commands-in-redis/3-5-sorted-sets/).
//
// The main different to other sets is, every value of set is associated with a score,
// that is used in order to take the sorted set ordered, from the smallest to the greatest score.
//
// The sorted set has O(log(N)) time complexity when doing Add(ZADD) and
// Remove(ZREM) operations and O(1) time complexity when doing Contains operations.
package zset

import (
	"math"
	"sync"
	"unsafe"

	"github.com/bytedance/gopkg/lang/fastrand"
)

// Float64RangeOpt describes the whether the min/max is exclusive in score range.
type Float64RangeOpt struct {
	ExcludeMin bool
	ExcludeMax bool
}

// Float64Node represents an element of Float64Set.
type Float64Node struct {
	Value string
	Score float64
}

// Float64Set is a sorted set implementation with string value and float64 score.
type Float64Set struct {
	mu   sync.RWMutex
	dict map[string]float64
	list *float64List
}

// NewFloat64 returns an empty string sorted set with int score.
// strings are sorted in ascending order.
func NewFloat64() *Float64Set {
	return &Float64Set{
		dict: make(map[string]float64),
		list: newFloat64List(),
	}
}

// UnionStoreFloat64 returns the union of given sorted sets, the resulting score of
// a value is the sum of its scores in the sorted sets where it exists.
//
// UnionStoreFloat64 is the replacement of UNIONSTORE command of redis.
func UnionStoreFloat64(zs ...*Float64Set) *Float64Set {
	dest := NewFloat64()
	for _, z := range zs {
		for _, n := range z.Range(0, z.Len()-1) {
			dest.Add(n.Score, n.Value)
		}
	}
	return dest
}

// InterStoreFloat64 returns the intersection of given sorted sets, the resulting
// score of a value is the sum of its scores in the sorted sets where it exists.
//
// InterStoreFloat64 is the replacement of INTERSTORE command of redis.
func InterStoreFloat64(zs ...*Float64Set) *Float64Set {
	dest := NewFloat64()
	if len(zs) == 0 {
		return dest
	}
	for _, n := range zs[0].Range(0, -1) {
		ok := true
		for _, z := range zs[1:] {
			if !z.Contains(n.Value) {
				ok = false
			}
		}
		if ok {
			dest.Add(n.Score, n.Value)
		}
	}
	return dest
}

// Len returns the length of Float64Set.
//
// Len is the replacement of ZCARD command of redis.
func (z *Float64Set) Len() int {
	z.mu.RLock()
	defer z.mu.RUnlock()

	return z.list.length
}

// Add adds a new value or update the score of an existing value.
// Returns true if the value is newly created.
//
// Add is the replacement of ZADD command of redis.
func (z *Float64Set) Add(score float64, value string) bool {
	z.mu.Lock()
	defer z.mu.Unlock()

	oldScore, ok := z.dict[value]
	if ok {
		// Update score if need
		if score != oldScore {
			_ = z.list.UpdateScore(oldScore, value, score)
			z.dict[value] = score
		}
		return false
	}

	// Insert a new element
	z.list.Insert(score, value)
	z.dict[value] = score
	return true
}

// Remove removes a value from the sorted set.
// Returns score of the removed value and true if the node was found and deleted,
// otherwise returns (0.0, false).
//
// Remove is the replacement of ZREM command of redis.
func (z *Float64Set) Remove(value string) (float64, bool) {
	z.mu.Lock()
	defer z.mu.Unlock()

	score, ok := z.dict[value]
	if !ok {
		return 0, false
	}
	delete(z.dict, value)
	z.list.Delete(score, value)
	return score, true
}

// IncrBy increments the score of value in the sorted set by incr.
// If value does not exist in the sorted set, it is added with incr as its score
// (as if its previous score was zero).
//
// IncrBy is the replacement of ZINCRBY command of redis.
func (z *Float64Set) IncrBy(incr float64, value string) (float64, bool) {
	z.mu.Lock()
	defer z.mu.Unlock()

	oldScore, ok := z.dict[value]
	if !ok {
		// Insert a new element
		z.list.Insert(incr, value)
		z.dict[value] = incr
		return incr, false
	} else {
		// Update score
		newScore := oldScore + incr
		_ = z.list.UpdateScore(oldScore, value, newScore)
		z.dict[value] = newScore
		return newScore, true
	}
}

// Contains returns whether the value exists in sorted set.
func (z *Float64Set) Contains(value string) bool {
	_, ok := z.Score(value)
	return ok
}

// Score returns the score of the value in the sorted set.
//
// Score is the replacement of ZSCORE command of redis.
func (z *Float64Set) Score(value string) (float64, bool) {
	z.mu.RLock()
	defer z.mu.RUnlock()

	score, ok := z.dict[value]
	return score, ok
}

// Rank returns the rank of element in the sorted set, with the scores
// ordered from low to high.
// The rank (or index) is 0-based, which means that the member with the lowest
// score has rank 0.
// -1 is returned when value is not found.
//
// Rank is the replacement of ZRANK command of redis.
func (z *Float64Set) Rank(value string) int {
	z.mu.RLock()
	defer z.mu.RUnlock()

	score, ok := z.dict[value]
	if !ok {
		return -1
	}
	// NOTE: list.Rank returns 1-based rank.
	return z.list.Rank(score, value) - 1
}

// RevRank returns the rank of element in the sorted set, with the scores
// ordered from high to low.
// The rank (or index) is 0-based, which means that the member with the highest
// score has rank 0.
// -1 is returned when value is not found.
//
// RevRank is the replacement of ZREVRANK command of redis.
func (z *Float64Set) RevRank(value string) int {
	z.mu.RLock()
	defer z.mu.RUnlock()

	score, ok := z.dict[value]
	if !ok {
		return -1
	}
	// NOTE: list.Rank returns 1-based rank.
	return z.list.Rank(score, value) - 1
}

// Count returns the number of elements in the sorted set at element with a score
// between min and max (including elements with score equal to min or max).
//
// Count is the replacement of ZCOUNT command of redis.
func (z *Float64Set) Count(min, max float64) int {
	return z.CountWithOpt(min, max, Float64RangeOpt{})
}

func (z *Float64Set) CountWithOpt(min, max float64, opt Float64RangeOpt) int {
	z.mu.RLock()
	defer z.mu.RUnlock()

	first := z.list.FirstInRange(min, max, opt)
	if first == nil {
		return 0
	}
	// sub 1 for 1-based rank
	firstRank := z.list.Rank(first.score, first.value) - 1
	last := z.list.LastInRange(min, max, opt)
	if last == nil {
		return z.list.length - firstRank
	}
	// sub 1 for 1-based rank
	lastRank := z.list.Rank(last.score, last.value) - 1
	return lastRank - firstRank + 1
}

// Range returns the specified inclusive range of elements in the sorted set by rank(index).
// Both start and stop are 0-based, they can also be negative numbers indicating
// offsets from the end of the sorted set, with -1 being the last element of the sorted set,
// and so on.
//
// The returned elements are ordered by score, from lowest to highest.
// Elements with the same score are ordered lexicographically.
//
// This function won't panic even when the given rank out of range.
//
// Range is the replacement of ZRANGE command of redis.
func (z *Float64Set) Range(start, stop int) []Float64Node {
	z.mu.RLock()
	defer z.mu.RUnlock()

	// Convert negative rank to positive
	if start < 0 {
		start = z.list.length + start
	}
	if stop < 0 {
		stop = z.list.length + stop
	}

	var res []Float64Node
	x := z.list.GetNodeByRank(start + 1) // 0-based rank -> 1-based rank
	for x != nil && start <= stop {
		start++
		res = append(res, Float64Node{
			Score: x.score,
			Value: x.value,
		})
		x = x.loadNext(0)
	}
	return res
}

// RangeByScore returns all the elements in the sorted set with a score
// between min and max (including elements with score equal to min or max).
// The elements are considered to be ordered from low to high scores.
//
// RangeByScore is the replacement of ZRANGEBYSCORE command of redis.
func (z *Float64Set) RangeByScore(min, max float64) []Float64Node {
	return z.RangeByScoreWithOpt(min, max, Float64RangeOpt{})
}

func (z *Float64Set) RangeByScoreWithOpt(min, max float64, opt Float64RangeOpt) []Float64Node {
	z.mu.RLock()
	defer z.mu.RUnlock()

	var res []Float64Node
	x := z.list.FirstInRange(min, max, opt)
	for x != nil && (x.score < max || (!opt.ExcludeMax && x.score == max)) {
		res = append(res, Float64Node{
			Score: x.score,
			Value: x.value,
		})
		x = x.loadNext(0)
	}
	return res
}

// RevRange returns the specified inclusive range of elements in the sorted set by rank(index).
// Both start and stop are 0-based, they can also be negative numbers indicating
// offsets from the end of the sorted set, with -1 being the first element of the sorted set,
// and so on.
//
// The returned elements are ordered by score, from highest to lowest.
// Elements with the same score are ordered in reverse lexicographical ordering.
//
// This function won't panic even when the given rank out of range.
//
// RevRange is the replacement of ZREVRANGE command of redis.
func (z *Float64Set) RevRange(start, stop int) []Float64Node {
	z.mu.RLock()
	defer z.mu.RUnlock()

	// Convert negative rank to positive
	if start < 0 {
		start = z.list.length + start
	}
	if stop < 0 {
		stop = z.list.length + stop
	}

	var res []Float64Node
	x := z.list.GetNodeByRank(z.list.length - start) // 0-based rank -> 1-based rank
	for x != nil && start <= stop {
		start++
		res = append(res, Float64Node{
			Score: x.score,
			Value: x.value,
		})
		x = x.prev
	}
	return res
}

// RevRangeByScore returns all the elements in the sorted set with a
// score between max and min (including elements with score equal to max or min).
// The elements are considered to be ordered from high to low scores.
//
// RevRangeByScore is the replacement of ZREVRANGEBYSCORE command of redis.
func (z *Float64Set) RevRangeByScore(max, min float64) []Float64Node {
	return z.RevRangeByScoreWithOpt(max, min, Float64RangeOpt{})
}

func (z *Float64Set) RevRangeByScoreWithOpt(max, min float64, opt Float64RangeOpt) []Float64Node {
	z.mu.RLock()
	defer z.mu.RUnlock()

	var res []Float64Node
	x := z.list.LastInRange(min, max, opt)
	for x != nil && (x.score > min || (!opt.ExcludeMin && x.score == min)) {
		res = append(res, Float64Node{
			Score: x.score,
			Value: x.value,
		})
		x = x.prev
	}
	return res
}

// RemoveRangeByRank removes all elements in the sorted set stored with rank
// between start and stop.
// Both start and stop are 0-based, they can also be negative numbers indicating
// offsets from the end of the sorted set, with -1 being the last element of the sorted set,
// and so on.
//
// RemoveRangeByRank is the replacement of ZREMRANGEBYRANK command of redis.
func (z *Float64Set) RemoveRangeByRank(start, stop int) []Float64Node {
	z.mu.RLock()
	defer z.mu.RUnlock()

	// Convert negative rank to positive
	if start < 0 {
		start = z.list.length + start
	}
	if stop < 0 {
		stop = z.list.length + stop
	}

	return z.list.DeleteRangeByRank(start+1, stop+1, z.dict) // 0-based rank -> 1-based rank
}

// RemoveRangeByScore removes all elements in the sorted set stored with a score
// between min and max (including elements with score equal to min or max).
//
// RemoveRangeByScore is the replacement of ZREMRANGEBYSCORE command of redis.
func (z *Float64Set) RemoveRangeByScore(min, max float64) []Float64Node {
	return z.RevRangeByScoreWithOpt(min, max, Float64RangeOpt{})
}

func (z *Float64Set) RemoveRangeByScoreWithOpt(min, max float64, opt Float64RangeOpt) []Float64Node {
	z.mu.RLock()
	defer z.mu.RUnlock()

	return z.list.DeleteRangeByScore(min, max, opt, z.dict)
}

//
// Skip list implementation.
//

const (
	maxLevel = 32   // same to ZSKIPLIST_MAXLEVEL
	p        = 0.25 // same to ZSKIPLIST_P
)

// float64ListNode is node of float64List.
type float64ListNode struct {
	score float64 // key for sorting, which is allowed to be repeated
	value string
	prev  *float64ListNode // back pointer that only available at level 1
	level int              // the length of optionalArray
	oparr optionalArray
}

func newFloat64ListNode(score float64, value string, level int) *float64ListNode {
	node := &float64ListNode{
		score: score,
		value: value,
		level: level,
	}
	if level > op1 {
		node.oparr.extra = new([op2]listLevel)
	}
	return node
}

func (n *float64ListNode) loadNext(i int) *float64ListNode {
	return (*float64ListNode)(n.oparr.loadNext(i))
}

func (n *float64ListNode) storeNext(i int, node *float64ListNode) {
	n.oparr.storeNext(i, unsafe.Pointer(node))
}

func (n *float64ListNode) loadSpan(i int) int {
	return n.oparr.loadSpan(i)
}

func (n *float64ListNode) storeSpan(i int, span int) {
	n.oparr.storeSpan(i, span)
}

func (n *float64ListNode) loadNextAndSpan(i int) (*float64ListNode, int) {
	return n.loadNext(i), n.loadSpan(i)
}

func (n *float64ListNode) storeNextAndSpan(i int, next *float64ListNode, span int) {
	n.storeNext(i, next)
	n.storeSpan(i, span)
}

func (n *float64ListNode) lessThan(score float64, value string) bool {
	if n.score < score {
		return true
	} else if n.score == score {
		return n.value < value
	}
	return false
}

func (n *float64ListNode) lessEqual(score float64, value string) bool {
	if n.score < score {
		return true
	} else if n.score == score {
		return n.value <= value
	}
	return false
}

func (n *float64ListNode) equal(score float64, value string) bool {
	return n.value == value && n.score == score
}

// float64List is a specialized skip list implementation for sorted set.
//
// It is almost implement the original
// algorithm described by William Pugh in " Lists: A Probabilistic
// Alternative to Balanced Trees", modified in three ways:
// a) this implementation allows for repeated scores.
// b) the comparison is not just by key (our 'score') but by satellite data(?).
// c) there is a back pointer, so it's a doubly linked list with the back
// pointers being only at "level 1". This allows to traverse the list
// from tail to head, useful for RevRange. */
type float64List struct {
	header       *float64ListNode
	tail         *float64ListNode
	length       int
	highestLevel int // highest level for now
}

func newFloat64List() *float64List {
	l := &float64List{
		header:       newFloat64ListNode(-math.MaxFloat64, "__HEADER", maxLevel), // FIXME:
		highestLevel: 1,
	}
	return l
}

// Insert inserts a new node in the skiplist. Assumes the element does not already
// exist (up to the caller to enforce that).
func (l *float64List) Insert(score float64, value string) *float64ListNode {
	var (
		update [maxLevel]*float64ListNode
		rank   [maxLevel + 1]int // +1 for eliminating a boundary judgment
	)

	x := l.header
	for i := l.highestLevel - 1; i >= 0; i-- {
		rank[i] = rank[i+1] // also fine when i == maxLevel - 1
		next := x.loadNext(i)
		for next != nil && next.lessThan(score, value) {
			rank[i] += x.loadSpan(i)
			x = next
			next = x.loadNext(i)
		}
		update[i] = x
	}

	// We assume the element is not already inside, since we allow duplicated
	// scores, reinserting the same element should never happen since the
	// caller of Add() should test in the hash table if the element is
	// already inside or not.
	level := l.randomLevel()
	if level > l.highestLevel {
		// create higher levels
		for i := l.highestLevel; i < level; i++ {
			rank[i] = 0
			update[i] = l.header
			update[i].storeSpan(i, l.length)
		}
		l.highestLevel = level
	}
	x = newFloat64ListNode(score, value, level)
	for i := 0; i < level; i++ {
		// update --> x --> update.next
		x.storeNext(i, update[i].loadNext(i))
		update[i].storeNext(i, x)
		// update[i].span is splitted to: new update[i].span and x.span
		x.storeSpan(i, update[i].loadSpan(i)-(rank[0]-rank[i]))
		update[i].storeSpan(i, (rank[0]-rank[i])+1)
	}
	// Increment span for untouched levels
	for i := level; i < l.highestLevel; i++ {
		update[i].storeSpan(i, update[i].loadSpan(i)+1)
	}

	// update back pointer
	if update[0] != l.header {
		x.prev = update[0]
	}

	if next := x.loadNext(0); next != nil { // not tail of skiplist
		next.prev = x
	} else {
		l.tail = x
	}
	l.length++

	return x
}

// randomLevel returns a level between (1, maxLevel] for insertion.
func (l *float64List) randomLevel() int {
	level := 1
	for fastrand.Uint32n(1/p) == 0 {
		level++
	}
	if level > maxLevel {
		return maxLevel
	}
	return level
}

// Rank finds the rank for an element by both score and value.
// Returns 0 when the element cannot be found, rank otherwise.
//
// NOTE: the rank is 1-based due to the span of l->header to the
// first element.
func (l *float64List) Rank(score float64, value string) int {
	rank := 0
	x := l.header
	for i := l.highestLevel - 1; i >= 0; i-- {
		next := x.loadNext(i)
		for next != nil && next.lessEqual(score, value) {
			rank += x.loadSpan(i)
			x = next
			next = x.loadNext(i)
		}

		// x might be equal to l->header, so test if obj is non-nil
		// TODO: Why not use if x != l.header?
		if x.equal(score, value) {
			return rank
		}
	}
	return 0
}

// deleteNode is a internal function for deleting node x in O(1) time by giving a
// update position matrix.
func (l *float64List) deleteNode(x *float64ListNode, update *[maxLevel]*float64ListNode) {
	for i := 0; i < l.highestLevel; i++ {
		if update[i].loadNext(i) == x {
			// Remove x, updaet[i].span = updaet[i].span + x.span - 1 (x removed)
			next, span := x.loadNextAndSpan(i)
			span += update[i].loadSpan(i) - 1
			update[i].storeNextAndSpan(i, next, span)
		} else {
			// x does not appear on this level, just update span
			update[i].storeSpan(i, update[i].loadSpan(i)-1)
		}
	}
	if next := x.loadNext(0); next != nil { // not tail of skiplist
		next.prev = x.prev
	} else {
		l.tail = x.prev
	}
	for l.highestLevel > 1 && l.header.loadNext(l.highestLevel-1) != nil {
		// Clear the pointer and span for safety
		l.header.storeNextAndSpan(l.highestLevel-1, nil, 0)
		l.highestLevel--
	}
	l.length--
}

// Delete deletes an element with matching score/element from the skiplist.
// The deleted node is returned if the node was found, otherwise 0 is returned.
func (l *float64List) Delete(score float64, value string) *float64ListNode {
	var update [maxLevel]*float64ListNode

	x := l.header
	for i := l.highestLevel - 1; i >= 0; i-- {
		next := x.loadNext(i)
		for next != nil && next.lessThan(score, value) {
			x = next
			next = x.loadNext(i)
		}
		update[i] = x
	}
	x = x.loadNext(0)
	if x != nil && x.equal(score, value) {
		l.deleteNode(x, &update)
		return x
	}
	// not found
	return nil
}

// UpdateScore updates the score of an element inside the sorted set skiplist.

// NOTE: the element must exist and must match 'score'.
// This function does not update the score in the hash table side, the
// caller should take care of it.
//
// NOTE: this function attempts to just update the node, in case after
// the score update, the node would be exactly at the same position.
// Otherwise the skiplist is modified by removing and re-adding a new
// element, which is more costly.
//
// The function returns the updated element skiplist node pointer.
func (l *float64List) UpdateScore(oldScore float64, value string, newScore float64) *float64ListNode {
	var update [maxLevel]*float64ListNode

	x := l.header
	for i := l.highestLevel - 1; i >= 0; i-- {
		next := x.loadNext(i)
		for next != nil && next.lessThan(oldScore, value) {
			x = next
			next = x.loadNext(i)
		}
		update[i] = x
	}

	// Jump to our element: note that this function assumes that the
	// element with the matching score exists.
	x = x.loadNext(0)

	// Fastpath: If the node, after the score update, would be still exactly
	// at the same position, we can just update the score without
	// actually removing and re-inserting the element in the skiplist.
	if next := x.loadNext(0); (x.prev == nil || x.prev.score < newScore) &&
		(next == nil || next.score > newScore) {
		x.score = newScore
		return x
	}

	// No way to reuse the old node: we need to remove and insert a new
	// one at a different place.
	v := x.value
	l.deleteNode(x, &update)
	newNode := l.Insert(newScore, v)
	return newNode
}

func greaterThanMin(value float64, min float64, ex bool) bool {
	if ex {
		return value > min
	} else {
		return value >= min
	}
}

func lessThanMax(value float64, max float64, ex bool) bool {
	if ex {
		return value < max
	} else {
		return value <= max
	}
}

// DeleteRangeByScore deletes all the elements with score between min and max
// from the skiplist.
// Both min and max can be inclusive or exclusive (see Float64RangeOpt).
// When inclusive a score >= min && score <= max is deleted.
//
// This function returns count of deleted elements.
func (l *float64List) DeleteRangeByScore(min, max float64, opt Float64RangeOpt, dict map[string]float64) []Float64Node {
	var (
		update  [maxLevel]*float64ListNode
		removed []Float64Node
	)

	x := l.header
	for i := l.highestLevel - 1; i >= 0; i-- {
		next := x.loadNext(i)
		for next != nil && !greaterThanMin(next.score, min, opt.ExcludeMin) {
			x = next
			next = x.loadNext(i)
		}
		update[i] = x
	}

	// Current node is the last with score not greater than min. */
	x = x.loadNext(0)

	// Delete nodes in range
	for x != nil && lessThanMax(x.score, max, opt.ExcludeMax) {
		next := x.loadNext(0)
		l.deleteNode(x, &update)
		delete(dict, x.value)
		removed = append(removed, Float64Node{
			Value: x.value,
			Score: x.score,
		})
		x = next
	}

	return removed
}

// Delete all the elements with rank between start and end from the skiplist.
// Start and end are inclusive.
//
// NOTE: start and end need to be 1-based
func (l *float64List) DeleteRangeByRank(start, end int, dict map[string]float64) []Float64Node {
	var (
		update    [maxLevel]*float64ListNode
		removed   []Float64Node
		traversed int
	)

	x := l.header
	for i := l.highestLevel - 1; i >= 0; i-- {
		next, span := x.loadNextAndSpan(i)
		for next != nil && traversed+span < start {
			traversed += span
			x = next
			next, span = x.loadNextAndSpan(i)
		}
		update[i] = x
	}

	traversed++
	x = x.loadNext(0)
	// Delete nodes in range
	for x != nil && traversed <= end {
		next := x.loadNext(0)
		l.deleteNode(x, &update)
		delete(dict, x.value)
		removed = append(removed, Float64Node{
			Value: x.value,
			Score: x.score,
		})
		traversed++
		x = next
	}
	return removed
}

// GetNodeByRank finds an element by its rank. The rank argument needs to be 1-based.
func (l *float64List) GetNodeByRank(rank int) *float64ListNode {
	var traversed int

	x := l.header
	for i := l.highestLevel - 1; i >= 0; i-- {
		next, span := x.loadNextAndSpan(i)
		for next != nil && traversed+span <= rank {
			traversed += span
			x = next
			next, span = x.loadNextAndSpan(i)
		}
		if traversed == rank {
			return x
		}
	}
	return nil
}

// FirstInRange finds the first node that is contained in the specified range.
func (l *float64List) FirstInRange(min, max float64, opt Float64RangeOpt) *float64ListNode {
	if !l.IsInRange(min, max, opt) {
		return nil
	}

	x := l.header
	for i := l.highestLevel - 1; i >= 0; i-- {
		next := x.loadNext(i)
		for next != nil && !greaterThanMin(next.score, min, opt.ExcludeMin) {
			x = next
			next = x.loadNext(i)
		}
	}

	// The next node MUST not be NULL (excluded by IsInRange)
	x = x.loadNext(0)
	if !lessThanMax(x.score, max, opt.ExcludeMax) {
		return nil
	}
	return x
}

// LastInRange finds the last node that is contained in the specified range.
func (l *float64List) LastInRange(min, max float64, opt Float64RangeOpt) *float64ListNode {
	if !l.IsInRange(min, max, opt) {
		return nil
	}

	x := l.header
	for i := l.highestLevel - 1; i >= 0; i-- {
		next := x.loadNext(i)
		for next != nil && lessThanMax(next.score, max, opt.ExcludeMax) {
			x = next
			next = x.loadNext(i)
		}
	}

	// The node x must not be NULL (excluded by IsInRange)
	if !greaterThanMin(x.score, min, opt.ExcludeMin) {
		return nil
	}
	return x
}

// IsInRange returns whether there is a port of sorted set in given range.
func (l *float64List) IsInRange(min, max float64, opt Float64RangeOpt) bool {
	// Test empty range
	if min > max || (min == max && (opt.ExcludeMin || opt.ExcludeMax)) {
		return false
	}
	if l.tail == nil || !greaterThanMin(l.tail.score, min, opt.ExcludeMin) {
		return false
	}
	if next := l.header.loadNext(0); next == nil || !lessThanMax(next.score, max, opt.ExcludeMax) {
		return false
	}
	return true
}
