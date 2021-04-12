# Asynccache

# Feature
1. 允许预设默认值
2. 支持过期淘汰(标记清理策略), 注意: 如开启淘汰机制, 可能会导致未使用的预设值被清理

# 优化refresh性能前后对比：

```text
benchmark                        old ns/op     new ns/op     delta
BenchmarkGet-12                  30.4          30.9          +1.64%
BenchmarkGetParallel-12          4.74          5.84          +23.21%
BenchmarkGetOrSet-12             56.1          55.9          -0.36%
BenchmarkGetOrSetParallel-12     11.7          12.1          +3.42%
BenchmarkRefresh-12              355           48.5          -86.34%
BenchmarkRefreshParallel-12      127           19.7          -84.49%

benchmark                        old allocs     new allocs     delta
BenchmarkGet-12                  0              0              +0.00%
BenchmarkGetParallel-12          0              0              +0.00%
BenchmarkGetOrSet-12             1              1              +0.00%
BenchmarkGetOrSetParallel-12     1              1              +0.00%
BenchmarkRefresh-12              5              0              -100.00%
BenchmarkRefreshParallel-12      5              0              -100.00%

benchmark                        old bytes     new bytes     delta
BenchmarkGet-12                  0             0             +0.00%
BenchmarkGetParallel-12          0             0             +0.00%
BenchmarkGetOrSet-12             16            16            +0.00%
BenchmarkGetOrSetParallel-12     16            16            +0.00%
BenchmarkRefresh-12              400           0             -100.00%
BenchmarkRefreshParallel-12      400           0             -100.00%
```
