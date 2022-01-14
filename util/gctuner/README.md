# gctuner

## Introduction

Inspired
by [How We Saved 70K Cores Across 30 Mission-Critical Services (Large-Scale, Semi-Automated Go GC Tuning @Uber)](https://eng.uber.com/how-we-saved-70k-cores-across-30-mission-critical-services/)
.

```text
 _______________  => limit: host/cgroup memory hard limit
|               |
|---------------| => gc_trigger: heap_live + heap_live * GCPercent / 100
|               |
|---------------|
|   heap_live   |
|_______________|
```

Go runtime trigger GC when hit `gc_trigger` which affected by `GCPercent` and `heap_live`.

Assuming that we have stable traffic, our application will always have 100 MB live heap, so the runtime will trigger GC
once heap hits 200 MB(by default GOGC=100). The heap size will be changed like: `100MB => 200MB => 100MB => 200MB => ...`.

But in real production, our application may have 4 GB memory resources, so no need to GC so frequently.

The gctuner helps to change the GOGC(GCPercent) dynamically at runtime, set the appropriate GCPercent according to current
memory usage.

### How it works

```text
 _______________  => limit: host/cgroup memory hard limit
|               |
|---------------| => threshold: increase GCPercent when gc_trigger < threshold
|               |
|---------------| => gc_trigger: heap_live + heap_live * GCPercent / 100
|               |
|---------------|
|   heap_live   |
|_______________|

threshold = inuse + inuse * (gcPercent / 100)
=> gcPercent = (threshold - inuse) / inuse * 100

if threshold < 2*inuse, so gcPercent < 100, and GC positively to avoid OOM
if threshold > 2*inuse, so gcPercent > 100, and GC negatively to reduce GC times
```

## Usage

The recommended threshold is 70% of the memory limit.

```go

// Get mem limit from the host machine or cgroup file.
limit := 4 * 1024 * 1024 * 1024
threshold := limit * 0.7

gctuner.Tuning(threshold)
```
