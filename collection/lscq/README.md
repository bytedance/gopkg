# LSCQ

LSCQ is a scalable, unbounded, multiple-producer and multiple-consumer FIFO queue in Go language. 

In the benchmark(AMD 3700x, running at 3.6 GHZ, -cpu=16), the LSCQ outperforms lock-based linked queue **5x ~ 6x** in most cases. Since built-in channel is a bounded queue, we can only compared it in EnqueueDequeuePair,  the LSCQ outperforms built-in channel **8x ~ 9x** in this case.

The ideas behind the LSCQ are [A Scalable, Portable, and Memory-Efficient Lock-Free FIFO Queue](https://arxiv.org/abs/1908.04511) and [Fast Concurrent Queues for x86 Processors](https://www.cs.tau.ac.il/~mad/publications/ppopp2013-x86queues.pdf).



## QuickStart

```go
package main

import (
	"github.com/bytedance/gopkg/collection/lscq"
)

func main() {
	  q := lscq.NewUint64()
	  q.Enqueue(100)
	  println(q.Dequeue())
}
```



## Benchmark

- Go version: go1.16.2 linux/amd64 
- OS: ubuntu 18.04
- CPU: AMD 3700x(8C16T), running at 3.6 GHZ (disable CPU turbo boost)
- MEMORY: 16G x 2 DDR4 memory, running at 3200 MHZ



### CPU=100

```bash
go test -bench=. -cpu=100 -run=NOTEST -benchtime=1000000x
```

![benchmarkcpu100](https://raw.githubusercontent.com/zhangyunhao116/public-data/master/lscq-benchmark-cpu100.png)

```
Default/EnqueueOnly/LSCQ-100            38.9ns ±14%
Default/EnqueueOnly/linkedQ-100          209ns ± 3%
Default/EnqueueOnly/msqueue-100          379ns ± 2%
Default/DequeueOnlyEmpty/LSCQ-100       10.0ns ±31%
Default/DequeueOnlyEmpty/linkedQ-100    79.2ns ± 4%
Default/DequeueOnlyEmpty/msqueue-100    7.59ns ±44%
Default/Pair/LSCQ-100                   58.7ns ± 7%
Default/Pair/linkedQ-100                 324ns ± 5%
Default/Pair/msqueue-100                 393ns ± 2%
Default/50Enqueue50Dequeue/LSCQ-100     34.9ns ± 8%
Default/50Enqueue50Dequeue/linkedQ-100   183ns ± 7%
Default/50Enqueue50Dequeue/msqueue-100   191ns ± 3%
Default/30Enqueue70Dequeue/LSCQ-100     78.5ns ± 4%
Default/30Enqueue70Dequeue/linkedQ-100   148ns ± 8%
Default/30Enqueue70Dequeue/msqueue-100   136ns ± 4%
Default/70Enqueue30Dequeue/LSCQ-100     36.2ns ±13%
Default/70Enqueue30Dequeue/linkedQ-100   195ns ± 4%
Default/70Enqueue30Dequeue/msqueue-100   267ns ± 2%
```



### CPU=16

```bash
go test -bench=. -cpu=16 -run=NOTEST -benchtime=1000000x
```

![benchmarkcpu16](https://raw.githubusercontent.com/zhangyunhao116/public-data/master/lscq-benchmark-cpu16.png)

```
Default/EnqueueOnly/LSCQ-16             33.7ns ± 5%
Default/EnqueueOnly/linkedQ-16           177ns ± 2%
Default/EnqueueOnly/msqueue-16           370ns ± 1%
Default/DequeueOnlyEmpty/LSCQ-16        3.27ns ±47%
Default/DequeueOnlyEmpty/linkedQ-16     91.1ns ± 2%
Default/DequeueOnlyEmpty/msqueue-16     3.23ns ±46%
Default/Pair/LSCQ-16                    56.1ns ± 3%
Default/Pair/linkedQ-16                  290ns ± 1%
Default/Pair/msqueue-16                  367ns ± 1%
Default/50Enqueue50Dequeue/LSCQ-16      31.8ns ± 3%
Default/50Enqueue50Dequeue/linkedQ-16    157ns ± 8%
Default/50Enqueue50Dequeue/msqueue-16    188ns ± 4%
Default/30Enqueue70Dequeue/LSCQ-16      73.8ns ± 2%
Default/30Enqueue70Dequeue/linkedQ-16    149ns ± 5%
Default/30Enqueue70Dequeue/msqueue-16    123ns ± 2%
Default/70Enqueue30Dequeue/LSCQ-16      28.8ns ± 4%
Default/70Enqueue30Dequeue/linkedQ-16    176ns ± 3%
Default/70Enqueue30Dequeue/msqueue-16    261ns ± 2%
```



### CPU=1

```bash
go test -bench=. -cpu=1 -run=NOTEST -benchtime=1000000x
```

![benchmarkcpu1](https://raw.githubusercontent.com/zhangyunhao116/public-data/master/lscq-benchmark-cpu1.png)

```
name                                    time/op
Default/EnqueueOnly/LSCQ                17.3ns ± 1%
Default/EnqueueOnly/linkedQ             59.9ns ± 6%
Default/EnqueueOnly/msqueue             67.1ns ± 2%
Default/DequeueOnlyEmpty/LSCQ           4.77ns ± 1%
Default/DequeueOnlyEmpty/linkedQ        11.3ns ± 2%
Default/DequeueOnlyEmpty/msqueue        3.14ns ± 1%
Default/Pair/LSCQ                       36.7ns ± 0%
Default/Pair/linkedQ                    56.2ns ± 6%
Default/Pair/msqueue                    60.2ns ± 2%
Default/50Enqueue50Dequeue/LSCQ         23.1ns ± 2%
Default/50Enqueue50Dequeue/linkedQ      34.1ns ± 3%
Default/50Enqueue50Dequeue/msqueue      40.8ns ± 9%
Default/30Enqueue70Dequeue/LSCQ         26.5ns ± 2%
Default/30Enqueue70Dequeue/linkedQ      27.0ns ±28%
Default/30Enqueue70Dequeue/msqueue      26.7ns ± 7%
Default/70Enqueue30Dequeue/LSCQ         25.2ns ± 5%
Default/70Enqueue30Dequeue/linkedQ      47.3ns ± 5%
Default/70Enqueue30Dequeue/msqueue      55.2ns ± 8%
```

