<p align="center">
  <img src="https://raw.githubusercontent.com/zhangyunhao116/public-data/master/skipset-logo2.png"/>
</p>

## Introduction

skipset is a high-performance concurrent set based on skip list. In typical pattern(100000 operations, 90%CONTAINS 9%ADD 1%REMOVE), the skipset up to 3x ~ 15x faster than the built-in sync.Map.

Different from the sync.Map, the items in the skipset are always sorted, and the `Contains` and `Range` operations are wait-free (A goroutine is guaranteed to complete an operation as long as it keeps taking steps, regardless of the activity of other goroutines).



## Features

- Concurrent safe API with high-performance.
- Wait-free Contains and Range operations.
- Sorted items.



## When should you use skipset

In these situations, `skipset` is better

- **Sorted elements is needed**.
- **Concurrent calls multiple operations**. such as use both `Contains` and `Add` at the same time.
- **Memory intensive**. The skipset save at least 50% memory in the benchmark.

In these situations, `sync.Map` is better

- Only one goroutine access the set for most of the time, such as insert a batch of elements and then use only `Contains` (use built-in map is even better).



## QuickStart

```go
package main

import (
	"fmt"

	"github.com/bytedance/gopkg/collection/skipset"
)

func main() {
	l := NewInt()

	for _, v := range []int{10, 12, 15} {
		if l.Add(v) {
			fmt.Println("skipset add", v)
		}
	}

	if l.Contains(10) {
		fmt.Println("skipset contains 10")
	}

	l.Range(func(value int) bool {
		fmt.Println("skipset range found ", value)
		return true
	})

	l.Remove(15)
	fmt.Printf("skipset contains %d items\r\n", l.Len())
}

```



## Benchmark

Go version: go1.16.2 linux/amd64

CPU: AMD 3700x(8C16T), running at 3.6GHz

OS: ubuntu 18.04

MEMORY: 16G x 2 (3200MHz)

![benchmark](https://raw.githubusercontent.com/zhangyunhao116/public-data/master/skipset-benchmark.png)

```shell
$ go test -run=NOTEST -bench=. -benchtime=100000x -benchmem -count=10 -timeout=60m  > x.txt
$ benchstat x.txt
```

```
name                                             time/op
Add/skipset-16                                    113ns ±11%
Add/sync.Map-16                                   682ns ± 5%
Contains100Hits/skipset-16                       13.7ns ±13%
Contains100Hits/sync.Map-16                      15.2ns ±17%
Contains50Hits/skipset-16                        14.6ns ±24%
Contains50Hits/sync.Map-16                       14.7ns ± 8%
ContainsNoHits/skipset-16                        14.3ns ± 5%
ContainsNoHits/sync.Map-16                       12.4ns ±15%
50Add50Contains/skipset-16                       61.3ns ± 7%
50Add50Contains/sync.Map-16                       572ns ± 7%
30Add70Contains/skipset-16                       49.2ns ±14%
30Add70Contains/sync.Map-16                       599ns ± 6%
1Remove9Add90Contains/skipset-16                 36.6ns ±17%
1Remove9Add90Contains/sync.Map-16                 496ns ± 6%
1Range9Remove90Add900Contains/skipset-16         41.8ns ±10%
1Range9Remove90Add900Contains/sync.Map-16        1.20µs ± 9%
StringAdd/skipset-16                              142ns ± 7%
StringAdd/sync.Map-16                             883ns ± 8%
StringContains50Hits/skipset-16                  21.2ns ±16%
StringContains50Hits/sync.Map-16                 20.8ns ± 5%
String30Add70Contains/skipset-16                 69.1ns ±18%
String30Add70Contains/sync.Map-16                 750ns ± 5%
String1Remove9Add90Contains/skipset-16           39.3ns ±12%
String1Remove9Add90Contains/sync.Map-16           619ns ± 3%
String1Range9Remove90Add900Contains/skipset-16   44.6ns ±17%
String1Range9Remove90Add900Contains/sync.Map-16  1.38µs ±10%

name                                             alloc/op
Add/skipset-16                                    58.0B ± 0%
Add/sync.Map-16                                    128B ± 0%
Contains100Hits/skipset-16                        0.00B     
Contains100Hits/sync.Map-16                       0.00B     
Contains50Hits/skipset-16                         0.00B     
Contains50Hits/sync.Map-16                        0.00B     
ContainsNoHits/skipset-16                         0.00B     
ContainsNoHits/sync.Map-16                        0.00B     
50Add50Contains/skipset-16                        29.0B ± 0%
50Add50Contains/sync.Map-16                       64.9B ± 6%
30Add70Contains/skipset-16                        17.0B ± 0%
30Add70Contains/sync.Map-16                       82.9B ±13%
1Remove9Add90Contains/skipset-16                  5.00B ± 0%
1Remove9Add90Contains/sync.Map-16                 55.6B ± 1%
1Range9Remove90Add900Contains/skipset-16          5.00B ± 0%
1Range9Remove90Add900Contains/sync.Map-16          302B ±10%
StringAdd/skipset-16                              90.0B ± 0%
StringAdd/sync.Map-16                              152B ± 0%
StringContains50Hits/skipset-16                   3.00B ± 0%
StringContains50Hits/sync.Map-16                  3.00B ± 0%
String30Add70Contains/skipset-16                  38.0B ± 0%
String30Add70Contains/sync.Map-16                 95.3B ± 5%
String1Remove9Add90Contains/skipset-16            22.0B ± 0%
String1Remove9Add90Contains/sync.Map-16           72.6B ± 2%
String1Range9Remove90Add900Contains/skipset-16    22.0B ± 0%
String1Range9Remove90Add900Contains/sync.Map-16    307B ±15%
```