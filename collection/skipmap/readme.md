<p align="center">
  <img src="https://raw.githubusercontent.com/zhangyunhao116/public-data/master/skipmap-logo.png"/>
</p>

## Introduction

skipmap is a high-performance concurrent map based on skip list. In typical pattern(one million operations, 90%LOAD 9%STORE 1%DELETE), the skipmap up to 3x ~ 10x faster than the built-in sync.Map.

Different from the sync.Map, the items in the skipmap are always sorted, and the `Load` and `Range` operations are wait-free (A goroutine is guaranteed to complete a operation as long as it keeps taking steps, regardless of the activity of other goroutines).



## Features

- Concurrent safe API with high-performance.
- Wait-free Load and Range operations.
- Sorted items.



## When should you use skipmap

In these situations, `skipmap` is better

- **Sorted elements is needed**.
- **Concurrent calls multiple operations**. such as use both `Load` and `Store` at the same time.

In these situations, `sync.Map` is better

- Only one goroutine access the map for most of the time, such as insert a batch of elements and then use only `Load` (use built-in map is even better).



## QuickStart

```go
package main

import (
	"fmt"

	"github.com/bytedance/gopkg/collection/skipmap"
)

func main() {
	l := skipmap.NewInt()

	for _, v := range []int{10, 12, 15} {
		l.Store(v, v+100)
	}

	v, ok := l.Load(10)
	if ok {
		fmt.Println("skipmap load 10 with value ", v)
	}

	l.Range(func(key int, value interface{}) bool {
		fmt.Println("skipmap range found ", key, value)
		return true
	})

	l.Delete(15)
	fmt.Printf("skipmap contains %d items\r\n", l.Len())
}

```



## Benchmark

Go version: go1.16.2 linux/amd64

CPU: AMD 3700x(8C16T), running at 3.6GHz

OS: ubuntu 18.04

MEMORY: 16G x 2 (3200MHz)

![benchmark](https://raw.githubusercontent.com/zhangyunhao116/public-data/master/skipmap-benchmark.png)

```shell
$ go test -run=NOTEST -bench=. -benchtime=100000x -benchmem -count=10 -timeout=60m  > x.txt
$ benchstat x.txt
```

```
name                                           time/op
Store/skipmap-16                                237ns ± 8%
Store/sync.Map-16                               676ns ± 5%
Load100Hits/skipmap-16                         13.2ns ±11%
Load100Hits/sync.Map-16                        14.7ns ±13%
Load50Hits/skipmap-16                          13.4ns ±16%
Load50Hits/sync.Map-16                         14.5ns ±22%
LoadNoHits/skipmap-16                          13.5ns ± 9%
LoadNoHits/sync.Map-16                         12.6ns ±16%
50Store50Load/skipmap-16                        132ns ± 3%
50Store50Load/sync.Map-16                       555ns ± 5%
30Store70Load/skipmap-16                       85.8ns ± 3%
30Store70Load/sync.Map-16                       577ns ± 5%
1Delete9Store90Load/skipmap-16                 46.4ns ±10%
1Delete9Store90Load/sync.Map-16                 494ns ± 5%
1Range9Delete90Store900Load/skipmap-16         53.0ns ± 8%
1Range9Delete90Store900Load/sync.Map-16        1.16µs ± 9%
StringStore/skipmap-16                          274ns ± 5%
StringStore/sync.Map-16                         876ns ± 5%
StringLoad50Hits/skipmap-16                    22.1ns ±14%
StringLoad50Hits/sync.Map-16                   18.9ns ±13%
String30Store70Load/skipmap-16                  118ns ±16%
String30Store70Load/sync.Map-16                 743ns ± 7%
String1Delete9Store90Load/skipmap-16           54.6ns ±13%
String1Delete9Store90Load/sync.Map-16           606ns ± 4%
String1Range9Delete90Store900Load/skipmap-16   62.1ns ±17%
String1Range9Delete90Store900Load/sync.Map-16  1.37µs ±21%

name                                           alloc/op
Store/skipmap-16                                 106B ± 0%
Store/sync.Map-16                                128B ± 0%
Load100Hits/skipmap-16                          0.00B     
Load100Hits/sync.Map-16                         0.00B     
Load50Hits/skipmap-16                           0.00B     
Load50Hits/sync.Map-16                          0.00B     
LoadNoHits/skipmap-16                           0.00B     
LoadNoHits/sync.Map-16                          0.00B     
50Store50Load/skipmap-16                        53.0B ± 0%
50Store50Load/sync.Map-16                       64.6B ± 2%
30Store70Load/skipmap-16                        32.0B ± 0%
30Store70Load/sync.Map-16                       74.7B ± 8%
1Delete9Store90Load/skipmap-16                  9.00B ± 0%
1Delete9Store90Load/sync.Map-16                 55.0B ± 0%
1Range9Delete90Store900Load/skipmap-16          9.00B ± 0%
1Range9Delete90Store900Load/sync.Map-16          295B ± 6%
StringStore/skipmap-16                           138B ± 0%
StringStore/sync.Map-16                          152B ± 0%
StringLoad50Hits/skipmap-16                     3.00B ± 0%
StringLoad50Hits/sync.Map-16                    3.00B ± 0%
String30Store70Load/skipmap-16                  52.0B ± 0%
String30Store70Load/sync.Map-16                 97.2B ±12%
String1Delete9Store90Load/skipmap-16            26.0B ± 0%
String1Delete9Store90Load/sync.Map-16           72.6B ± 2%
String1Range9Delete90Store900Load/skipmap-16    26.0B ± 0%
String1Range9Delete90Store900Load/sync.Map-16    309B ±28%
```