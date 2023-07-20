package ptuner

import (
	"fmt"
	"runtime/metrics"
)

const (
	schedLatencyMetricName = "/sched/latencies:seconds"
	cpuTotalMetricName     = "/cpu/classes/total:cpu-seconds"
	cpuUserMetricName      = "/cpu/classes/user:cpu-seconds"
	cpuIdleMetricName      = "/cpu/classes/idle:cpu-seconds"
	cpuGCMetricName        = "/cpu/classes/gc/total:cpu-seconds"
)

func newRuntimeAnalyzer() *runtimeAnalyzer {
	ra := &runtimeAnalyzer{
		lastStat: readRuntimeStat(),
	}
	return ra
}

type runtimeAnalyzer struct {
	lastStat runtimeStat
}

func (r *runtimeAnalyzer) Analyze() (schedLatency, cpuPercent float64) {
	stat := readRuntimeStat()
	lastStat := r.lastStat
	r.lastStat = stat

	// sched avg
	schedLatency += stat.latencyTotal - lastStat.latencyTotal

	// cpu avg
	total := stat.cpuTotal - lastStat.cpuTotal
	idle := stat.cpuIdle - lastStat.cpuIdle
	if total > 0 {
		cpuPercent = (total - idle) / total
	}
	return schedLatency, cpuPercent
}

type runtimeStat struct {
	latencyAvg, latencyP50, latencyP90, latencyP99, latencyTotal float64 // seconds
	cpuTotal, cpuUser, cpuGC, cpuIdle                            float64 // seconds
}

func (stat runtimeStat) String() string {
	ms := float64(1000)
	return fmt.Sprintf(
		"latency_avg=%.2fms latency_p50=%.2fms latency_p90=%.2fms latency_p99=%.2fms | "+
			"cpu_total=%.2fs cpu_user=%.2fs cpu_gc=%.2fs cpu_idle=%.2fs",
		stat.latencyAvg*ms, stat.latencyP50*ms, stat.latencyP90*ms, stat.latencyP99*ms,
		stat.cpuTotal, stat.cpuUser, stat.cpuGC, stat.cpuIdle,
	)
}

func readRuntimeStat() runtimeStat {
	var metricSamples = []metrics.Sample{
		{Name: schedLatencyMetricName},
		{Name: cpuTotalMetricName},
		{Name: cpuUserMetricName},
		{Name: cpuGCMetricName},
		{Name: cpuIdleMetricName},
	}
	metrics.Read(metricSamples)

	var stat runtimeStat
	stat.latencyAvg, stat.latencyP50, stat.latencyP90, stat.latencyP99, _, stat.latencyTotal = calculateSchedLatency(metricSamples[0])
	stat.cpuTotal, stat.cpuUser, stat.cpuGC, stat.cpuIdle = calculateCPUSeconds(
		metricSamples[1], metricSamples[2], metricSamples[3], metricSamples[4],
	)
	return stat
}

func calculateCPUSeconds(totalSample, userSample, gcSample, idleSample metrics.Sample) (total, user, gc, idle float64) {
	total = totalSample.Value.Float64()
	user = userSample.Value.Float64()
	gc = gcSample.Value.Float64()
	idle = idleSample.Value.Float64()
	return
}

func calculateSchedLatency(sample metrics.Sample) (avg, p50, p90, p99, max, total float64) {
	var (
		histogram  = sample.Value.Float64Histogram()
		totalCount uint64
		latestIdx  int
	)

	// range counts
	for idx, count := range histogram.Counts {
		if count > 0 {
			latestIdx = idx
		}
		totalCount += count
	}
	p50Count := totalCount / 2
	p90Count := uint64(float64(totalCount) * 0.90)
	p99Count := uint64(float64(totalCount) * 0.99)

	// range buckets
	var ranged uint64
	for idx, count := range histogram.Counts {
		if count == 0 {
			continue
		}
		ranged += count
		latency := histogram.Buckets[idx]
		total += latency * float64(count)
		if p99 == 0 && ranged >= p99Count {
			p99 = latency
		} else if p90 == 0 && ranged >= p90Count {
			p90 = latency
		} else if p50 == 0 && ranged >= p50Count {
			p50 = latency
		}
	}
	avg = total / float64(totalCount)
	max = histogram.Buckets[latestIdx]
	return
}
