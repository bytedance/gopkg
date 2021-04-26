# timex

Package timex provides an alternative to standard API time.Now(), which is much faster but less precise on linux amd64.

`timex` uses the `CLOCK_REALTIME_COARSE` and `CLOCK_MONOTONIC_COARSE` to get wall time and monotonic time, which is about 2.5x faster than the standard time library.
```
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
BenchmarkNow
BenchmarkNow-8      	69521247	        14.65 ns/op
BenchmarkStdNow
BenchmarkStdNow-8   	31656583	        37.26 ns/op
```

`timex` also provides high-resolution timestamp support for amd64 platforms, based on RDTSC instruction.
```
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
BenchmarkFenceRDTSC
BenchmarkFenceRDTSC-8   	99547878	        10.44 ns/op
BenchmarkRDTSC
BenchmarkRDTSC-8        	193517694	         6.215 ns/op
BenchmarkRDTSCP
BenchmarkRDTSCP-8       	122545428	         9.806 ns/op
```

