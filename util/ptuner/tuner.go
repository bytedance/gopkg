package ptuner

import (
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	defaultTuningFrequency = time.Minute
)

type Option func(t *tuner)

func WithMaxProcs(maxProcs int) Option {
	return func(t *tuner) {
		t.maxProcs = maxProcs
	}
}

func WithMinProcs(minProcs int) Option {
	return func(t *tuner) {
		t.minProcs = minProcs
	}
}

func WithTuningFrequency(duration time.Duration) Option {
	return func(t *tuner) {
		t.tuningFrequency = duration
	}
}

func WithTuningLimit(limit int) Option {
	return func(t *tuner) {
		t.tuningLimit = limit
	}
}

var tuningOnce int32

func Tuning(opts ...Option) error {
	if atomic.AddInt32(&tuningOnce, 1) > 1 {
		return fmt.Errorf("you can only tuning once")
	}
	t := new(tuner)
	t.tuningFrequency = defaultTuningFrequency
	t.oriProcs = runtime.GOMAXPROCS(0)
	t.curProcs = t.oriProcs
	t.minProcs = t.oriProcs
	if t.minProcs <= 0 {
		t.minProcs = 1
	}
	t.maxProcs = t.oriProcs * 2
	for _, opt := range opts {
		opt(t)
	}
	if t.tuningFrequency < time.Second {
		return fmt.Errorf("tuningFrequency should >= 1s")
	}
	t.schedPerf = make([]float64, t.maxProcs+1)
	t.cpuPerf = make([]float64, t.maxProcs+1)
	t.ra = newRuntimeAnalyzer()
	go t.tuning()
	return nil
}

type tuner struct {
	oriProcs        int // original GOMAXPROCS before turning
	curProcs        int // current GOMAXPROCS
	minProcs        int // as same as oriProcs by default
	maxProcs        int // as same as oriProcs*2 by default
	lastProcs       int // last GOMAXPROCS before change
	tuningFrequency time.Duration
	tuningLimit     int
	// runtime data
	ra        *runtimeAnalyzer
	schedPerf []float64
	cpuPerf   []float64
}

func (t *tuner) tuning() {
	log.Printf("PTuning: start tuning with min_procs=%d max_procs=%d", t.minProcs, t.maxProcs)

	ticker := time.NewTicker(t.tuningFrequency)
	defer ticker.Stop()
	tuned := 0
	loop := 0
	reportInterval := int(10 * time.Minute / t.tuningFrequency)
	for range ticker.C {
		loop++
		if t.tuningLimit != 0 && tuned >= t.tuningLimit {
			log.Printf("PTuning: hit tunning limit[%d], exit", t.tuningLimit)
			return
		}
		if loop%reportInterval == 0 {
			t.report()
		}

		newProcs := t.adjustProcs()
		if t.curProcs != newProcs {
			log.Printf("PTuning: change GOMAXPROCS from %d to %d", t.curProcs, newProcs)
			runtime.GOMAXPROCS(newProcs)
			t.lastProcs = t.curProcs
			t.curProcs = newProcs
			tuned++
		}
	}
}

func (t *tuner) adjustProcs() int {
	// save current perf data
	schedLatency, cpuPercent := t.ra.Analyze()
	t.schedPerf[t.curProcs] = schedLatency
	t.cpuPerf[t.curProcs] = cpuPercent
	log.Printf("PTuning: runtime analyze: sched_latency=%.6fms, cpu_percent=%.2f%%", schedLatency*1000, cpuPercent*100)

	if t.lastProcs > 0 && t.lastProcs != t.curProcs {
		// evaluate the turning effect
		if cpuPercent-t.cpuPerf[t.lastProcs] >= 0.05 {
			// if tuning cause too many cpu cost, so fallback to last procs
			return t.lastProcs
		}
	}

	// in the beginning, next pnumber always have the best performance data
	if t.curProcs < t.maxProcs && schedLatency > t.schedPerf[t.curProcs+1] {
		return t.curProcs + 1
	}
	// if prev pnumber have better performance, change to prev pnumber
	if t.curProcs > t.minProcs && schedLatency > t.schedPerf[t.curProcs-1] {
		return t.curProcs - 1
	}
	// if current pnumber is the minimum latency choice, check if there is the best choice
	bestP := t.bestSchedProc()
	// even we can guess the best pnumber, we still need to change it slowly
	if bestP > t.curProcs && t.curProcs < t.maxProcs {
		return t.curProcs + 1
	}
	if bestP < t.curProcs && t.curProcs > t.minProcs {
		return t.curProcs - 1
	}
	return t.curProcs
}

func (t *tuner) bestSchedProc() int {
	bestP := t.curProcs
	bestLatency := t.schedPerf[t.curProcs]
	for pn := t.minProcs; pn <= t.maxProcs; pn++ {
		if t.schedPerf[pn] < bestLatency {
			bestP = pn
			bestLatency = t.schedPerf[pn]
		}
	}
	return bestP
}

func (t *tuner) bestCPUProc() int {
	bestP := t.curProcs
	bestCPU := t.cpuPerf[t.curProcs]
	for pn := t.minProcs; pn <= t.maxProcs; pn++ {
		if t.cpuPerf[pn] < bestCPU {
			bestP = pn
			bestCPU = t.cpuPerf[pn]
		}
	}
	return bestP
}

func (t *tuner) report() {
	for pn := t.minProcs; pn <= t.maxProcs; pn++ {
		log.Printf("PTuning: reporting pnumber=%d sched_latency=%.6fms, cpu_percent=%.2f%%",
			pn, t.schedPerf[pn]*1000, t.cpuPerf[pn]*100,
		)
	}
}
