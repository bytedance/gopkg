// Copyright 2022 ByteDance Inc.
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

package gctuner

import (
	"math"
	"os"
	"runtime/debug"
	"strconv"
	"sync/atomic"
)

var (
	maxGCPercent uint32 = 500
	minGCPercent uint32 = 50
)

var defaultGCPercent uint32 = 100

func init() {
	gogcEnv := os.Getenv("GOGC")
	gogc, err := strconv.ParseInt(gogcEnv, 10, 32)
	if err != nil {
		return
	}
	defaultGCPercent = uint32(gogc)
}

// Tuning sets the threshold of heap which will be respect by gc tuner.
// When Tuning, the env GOGC will not be take effect.
// threshold: disable tuning if threshold == 0
func Tuning(threshold uint64) {
	// disable gc tuner if percent is zero
	if threshold <= 0 && globalTuner != nil {
		globalTuner.stop()
		globalTuner = nil
		return
	}

	if globalTuner == nil {
		globalTuner = newTuner(threshold)
		return
	}
	globalTuner.setThreshold(threshold)
}

// GetGCPercent returns the current GCPercent.
func GetGCPercent() uint32 {
	if globalTuner == nil {
		return defaultGCPercent
	}
	return globalTuner.getGCPercent()
}

// GetMaxGCPercent returns the max gc percent value.
func GetMaxGCPercent() uint32 {
	return atomic.LoadUint32(&maxGCPercent)
}

// SetMaxGCPercent sets the new max gc percent value.
func SetMaxGCPercent(n uint32) uint32 {
	return atomic.SwapUint32(&maxGCPercent, n)
}

// GetMinGCPercent returns the min gc percent value.
func GetMinGCPercent() uint32 {
	return atomic.LoadUint32(&minGCPercent)
}

// SetMinGCPercent sets the new min gc percent value.
func SetMinGCPercent(n uint32) uint32 {
	return atomic.SwapUint32(&minGCPercent, n)
}

// only allow one gc tuner in one process
var globalTuner *tuner = nil

/* Heap
 _______________  => limit: host/cgroup memory hard limit
|               |
|---------------| => threshold: increase GCPercent when gc_trigger < threshold
|               |
|---------------| => gc_trigger: heap_live + heap_live * GCPercent / 100
|               |
|---------------|
|   heap_live   |
|_______________|

Go runtime only trigger GC when hit gc_trigger which affected by GCPercent and heap_live.
So we can change GCPercent dynamically to tuning GC performance.
*/
type tuner struct {
	finalizer *finalizer
	gcPercent uint32
	threshold uint64 // high water level, in bytes
}

// tuning check the memory inuse and tune GC percent dynamically.
// Go runtime ensure that it will be called serially.
func (t *tuner) tuning() {
	inuse := readMemoryInuse()
	threshold := t.getThreshold()
	// stop gc tuning
	if threshold <= 0 {
		return
	}
	t.setGCPercent(calcGCPercent(inuse, threshold))
	return
}

// threshold = inuse + inuse * (gcPercent / 100)
// => gcPercent = (threshold - inuse) / inuse * 100
// if threshold < inuse*2, so gcPercent < 100, and GC positively to avoid OOM
// if threshold > inuse*2, so gcPercent > 100, and GC negatively to reduce GC times
func calcGCPercent(inuse, threshold uint64) uint32 {
	// invalid params
	if inuse == 0 || threshold == 0 {
		return defaultGCPercent
	}
	// inuse heap larger than threshold, use min percent
	if threshold <= inuse {
		return minGCPercent
	}
	gcPercent := uint32(math.Floor(float64(threshold-inuse) / float64(inuse) * 100))
	if gcPercent < minGCPercent {
		return minGCPercent
	} else if gcPercent > maxGCPercent {
		return maxGCPercent
	}
	return gcPercent
}

func newTuner(threshold uint64) *tuner {
	t := &tuner{
		gcPercent: defaultGCPercent,
		threshold: threshold,
	}
	t.finalizer = newFinalizer(t.tuning) // start tuning
	return t
}

func (t *tuner) stop() {
	t.finalizer.stop()
}

func (t *tuner) setThreshold(threshold uint64) {
	atomic.StoreUint64(&t.threshold, threshold)
}

func (t *tuner) getThreshold() uint64 {
	return atomic.LoadUint64(&t.threshold)
}

func (t *tuner) setGCPercent(percent uint32) uint32 {
	atomic.StoreUint32(&t.gcPercent, percent)
	return uint32(debug.SetGCPercent(int(percent)))
}

func (t *tuner) getGCPercent() uint32 {
	return atomic.LoadUint32(&t.gcPercent)
}
