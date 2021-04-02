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

package rand

import (
	impl "math/rand"
	"sync"
	"time"

	pid "github.com/choleraehyq/pid"
)

type lockedSource struct {
	_ [cacheLineSize]byte
	sync.Mutex
	*impl.Rand
}

func (ls *lockedSource) ExpFloat64() (n float64) {
	ls.Lock()
	n = ls.Rand.ExpFloat64()
	ls.Unlock()
	return
}

func (ls *lockedSource) NormFloat64() (n float64) {
	ls.Lock()
	n = ls.Rand.NormFloat64()
	ls.Unlock()
	return
}

func (ls *lockedSource) Seed(seed int64) {
	ls.Lock()
	ls.Rand.Seed(seed)
	ls.Unlock()
}

func (ls *lockedSource) Int63() (r int64) {
	ls.Lock()
	r = ls.Rand.Int63()
	ls.Unlock()
	return
}

func (ls *lockedSource) Uint32() (r uint32) {
	ls.Lock()
	r = ls.Rand.Uint32()
	ls.Unlock()
	return
}

func (ls *lockedSource) Uint64() (r uint64) {
	ls.Lock()
	r = ls.Rand.Uint64()
	ls.Unlock()
	return
}

func (ls *lockedSource) Int31() (r int32) {
	ls.Lock()
	r = ls.Rand.Int31()
	ls.Unlock()
	return
}

func (ls *lockedSource) Int() (r int) {
	ls.Lock()
	r = ls.Rand.Int()
	ls.Unlock()
	return
}

func (ls *lockedSource) Int63n(n int64) (r int64) {
	ls.Lock()
	r = ls.Rand.Int63n(n)
	ls.Unlock()
	return
}

func (ls *lockedSource) Int31n(n int32) (r int32) {
	ls.Lock()
	r = ls.Rand.Int31n(n)
	ls.Unlock()
	return
}

func (ls *lockedSource) Intn(n int) (r int) {
	ls.Lock()
	r = ls.Rand.Intn(n)
	ls.Unlock()
	return
}

func (ls *lockedSource) Float64() (r float64) {
	ls.Lock()
	r = ls.Rand.Float64()
	ls.Unlock()
	return
}

func (ls *lockedSource) Float32() (r float32) {
	ls.Lock()
	r = ls.Rand.Float32()
	ls.Unlock()
	return
}

func (ls *lockedSource) Perm(n int) (r []int) {
	ls.Lock()
	r = ls.Rand.Perm(n)
	ls.Unlock()
	return
}

func (ls *lockedSource) Shuffle(n int, swap func(i, j int)) {
	ls.Lock()
	ls.Rand.Shuffle(n, swap)
	ls.Unlock()
}

func (ls *lockedSource) Read(p []byte) (n int, err error) {
	ls.Lock()
	n, err = ls.Rand.Read(p)
	ls.Unlock()
	return
}

// Locked is p-sharded, can be used safely in situations where GOMAXPROCS is adjusted at runtime.
type Locked []*lockedSource

func (l Locked) ExpFloat64() float64 {
	return l[pid.GetPid()%shardsLen].ExpFloat64()
}

func (l Locked) NormFloat64() float64 {
	return l[pid.GetPid()%shardsLen].NormFloat64()
}

func (l Locked) Seed(seed int64) {
	l[pid.GetPid()%shardsLen].Seed(seed)
}

func (l Locked) Int63() int64 {
	return l[pid.GetPid()%shardsLen].Int63()
}

func (l Locked) Uint32() uint32 {
	return l[pid.GetPid()%shardsLen].Uint32()
}

func (l Locked) Uint64() uint64 {
	return l[pid.GetPid()%shardsLen].Uint64()
}

func (l Locked) Int31() int32 {
	return l[pid.GetPid()%shardsLen].Int31()
}

func (l Locked) Int() int {
	return l[pid.GetPid()%shardsLen].Int()
}

func (l Locked) Int63n(n int64) int64 {
	return l[pid.GetPid()%shardsLen].Int63n(n)
}

func (l Locked) Int31n(n int32) int32 {
	return l[pid.GetPid()%shardsLen].Int31n(n)
}

func (l Locked) Intn(n int) int {
	return l[pid.GetPid()%shardsLen].Intn(n)
}

func (l Locked) Float64() float64 {
	return l[pid.GetPid()%shardsLen].Float64()
}

func (l Locked) Float32() float32 {
	return l[pid.GetPid()%shardsLen].Float32()
}

func (l Locked) Perm(n int) []int {
	return l[pid.GetPid()%shardsLen].Perm(n)
}

func (l Locked) Shuffle(n int, swap func(i, j int)) {
	l[pid.GetPid()%shardsLen].Shuffle(n, swap)
}

func (l Locked) Read(p []byte) (int, error) {
	return l[pid.GetPid()%shardsLen].Read(p)
}

func NewLocked() Locked {
	s := make([]*lockedSource, shardsLen)
	for i := 0; i < shardsLen; i++ {
		s[i] = &lockedSource{
			Rand: impl.New(impl.NewSource(time.Now().UnixNano())),
		}
	}
	return s
}

func ExpFloat64() float64 {
	return defaultLocked.ExpFloat64()
}

func NormFloat64() float64 {
	return defaultLocked.NormFloat64()
}

func Seed(seed int64) {
	defaultLocked.Seed(seed)
}

func Int63() int64 {
	return defaultLocked.Int63()
}

func Uint32() uint32 {
	return defaultLocked.Uint32()
}

func Uint64() uint64 {
	return defaultLocked.Uint64()
}

func Int31() int32 {
	return defaultLocked.Int31()
}

func Int() int {
	return defaultLocked.Int()
}

func Int63n(n int64) int64 {
	return defaultLocked.Int63n(n)
}

func Int31n(n int32) int32 {
	return defaultLocked.Int31n(n)
}

func Intn(n int) int {
	return defaultLocked.Intn(n)
}

func Float64() float64 {
	return defaultLocked.Float64()
}

func Float32() float32 {
	return defaultLocked.Float32()
}

func Perm(n int) []int {
	return defaultLocked.Perm(n)
}

func Shuffle(n int, swap func(i, j int)) {
	defaultLocked.Shuffle(n, swap)
}

func Read(p []byte) (int, error) {
	return defaultLocked.Read(p)
}
