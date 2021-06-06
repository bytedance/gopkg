cd avo && go build && ./avo -avx2 -out ../avx2.s && ./avo -sse2 -out ../sse2.s && cd .. && go test -v

# XXH3 hash algorithm
A Go implementation of the 64/128 bit xxh3 algorithm. The original repository is here: [https://github.com/Cyan4973/xxHash].

Uses [https://github.com/zeebo/xxh3] as reference for SIMD implementation with optimization.

## Overview
For the input length larger than 240, xxh3 algorithm goes along with following steps to get the hash result.

### step1.  Initialize 8 accumulators used to store the middle result of each iterator.
```
xacc[0] = prime32_3
xacc[1] = prime64_1
xacc[2] = prime64_2
xacc[3] = prime64_3
xacc[4] = prime64_4
xacc[5] = prime32_2
xacc[6] = prime64_5
xacc[7] = prime32_1
```

### step2.  Process 1024 bytes of input data as one block each time
```
while remaining length > 1024{
    for i:=0, j:=0; i < 1024; i += 64, j+=8 {
        for n:=0; n<8; n++{
            inputN := input[i+8*n:i+8*n+8]
            secretN := inputN ^ secret[j+8*n:j+8*n+8]
            
            xacc[n^1] += inputN
            xacc[n]   +=  (secretN & 0xFFFFFFFF) * (secretN >> 32)
        }
    }
    
    xacc[n]   ^= xacc[n] >> 47
    xacc[n]   ^= secret[128+8*n:128+8*n:+8]
    xacc[n]   *= prime32_1
}
```

### step3.  Process remaining stripes (total 1024 bytes at most)
```

for i:=0, j:=0; i < length; i += 64, j+=8 {
    for n:=0; n<8; n++{
        inputN := input[i+8*n:i+8*n+8]
        secretN := inputN ^ secret[j+8*n:j+8*n+8]
    
        xacc[n^1] += inputN
        xacc[n]   += (secretN & 0xFFFFFFFF) * (secretN >> 32)
    }
}
```

### step4.  Process last stripe  (64 bytes at most)
```
for n:=0; n<8; n++{
    inputN := input[(length-64): (length-64)+8]
    secretN := inputN ^ secret[121+8*n, 121+8*n+8]

    xacc[n^1] += inputN
    xacc[n]   += (secretN & 0xFFFFFFFF) * (secretN >> 32)
}
```

### step5.  Merge & Avalanche accumulators
```
acc = length * prime64_1
acc += mix(xacc[0]^secret11, xacc[1]^secret19) + mix(xacc[2]^secret27, xacc[3]^secret35) +
    mix(xacc[4]^secret43, xacc[5]^secret51) + mix(xacc[6]^secret59, xacc[7]^secret67)

acc ^= acc >> 37
acc *= 0x165667919e3779f9
acc ^= acc >> 32
```


## Quickstart
```
package main

import "github.com/bytedance/gopkg/util/xxhash3"

func main() {
	println(xxhash3.HashString("hello world!"))
}
```
## Benchmark
go version: go1.15.10 linux/amd64
CPU: Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz
OS: Linux bluehorse 5.8.0-48-generic #54~20.04.1-Ubuntu SMP
MEMORY: 32G

```
go test -run=None -bench=. -cpu=4 -benchtime=1000x -count=10 > 1000_10.txt && benchstat 1000_10.txt
```
```
name                               time/op
Default/Len0_16/Baseline64-4       88.4ns ± 2%
Default/Len0_16/Target64-4         88.6ns ± 0%
Default/Len0_16/Baseline128-4      182ns ± 0%
Default/Len0_16/Target128-4        176ns ± 0%
Default/Len17_128/Baseline64-4     1.09µs ± 2%
Default/Len17_128/Target64-4       1.07µs ± 2%
Default/Len17_128/Baseline128-4    1.91µs ± 0%
Default/Len17_128/Target128-4      1.76µs ± 1%
Default/Len129_240/Baseline64-4    1.93µs ± 1%
Default/Len129_240/Target64-4      1.89µs ± 2%
Default/Len129_240/Baseline128-4   3.01µs ± 0%
Default/Len129_240/Target128-4     2.82µs ± 3%
Default/Len241_1024/Baseline64-4   55.5µs ± 1%
Default/Len241_1024/Target64-4     47.9µs ± 0%
Default/Len241_1024/Baseline128-4  59.9µs ± 1%
Default/Len241_1024/Target128-4    52.8µs ± 1%
Default/Scalar/Baseline64-4        3.87ms ± 1%
Default/Scalar/Target64-4          3.52ms ± 2%
Default/Scalar/Baseline128-4       3.87ms ± 1%
Default/Scalar/Target128-4         3.52ms ± 1%
Default/AVX2/Baseline64-4          2.10ms ± 3%
Default/AVX2/Target64-4            1.93ms ± 2%
Default/AVX2/Baseline128-4         2.05ms ± 2%
Default/AVX2/Target128-4           1.91ms ± 1%
Default/SSE2/Baseline64-4          2.68ms ± 4%
Default/SSE2/Target64-4            2.61ms ± 2%
Default/SSE2/Baseline128-4         2.64ms ± 1%
Default/SSE2/Target128-4           2.63ms ± 4%
```