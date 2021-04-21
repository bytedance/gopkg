# timex

`timex` provides a `time.Now` function which is much faster and less precise.

`timex` uses the `CLOCK_REALTIME_COARSE` and `CLOCK_MONOTONIC_COARSE` to get wall time and monotonic time, which is about 2.5x faster than the standard time library.
```
cpu: Intel(R) Core(TM) i9-9980HK CPU @ 2.40GHz
BenchmarkNow
BenchmarkNow-8      	69521247	        14.65 ns/op
BenchmarkStdNow
BenchmarkStdNow-8   	31656583	        37.26 ns/op
```

`timex` also provides RDTSC support on amd64 platform.
