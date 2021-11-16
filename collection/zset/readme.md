# zset

[![Go Reference](https://pkg.go.dev/badge/github.com/bytedance/gopkg/collection/zset.svg)](https://pkg.go.dev/github.com/bytedance/gopkg/collection/zset)

## Introduction

zset provides a concurrent-safety sorted set, can be used as a local replacement of [Redis' zset](https://redis.com/ebook/part-2-core-concepts/chapter-3-commands-in-redis/3-5-sorted-sets/).

The main difference to other sets is, every value of set is associated with a score, that is used to take the sorted set ordered, from the smallest to the greatest score.

The zset has `O(log(N))` time complexity when doing Add(ZADD) and Remove(ZREM) operations and `O(1)` time complexity when doing Contains operations.

## Features

- Concurrent safe API
- Values are sorted with score
- Implementation equivalent to redis 
- Fast skiplist level randomization

## Comparison

| Redis command         | Go function         |
|-----------------------|---------------------|
| ZADD                  | Add                 |
| ZINCRBY               | IncrBy              |
| ZREM                  | Remove              |
| ZREMRANGEBYSCORE      | RemoveRangeByScore  |
| ZREMRANGEBYRANK       | RemoveRangeByRank   |
| ZUNION                | Union               |
| ZINTER                | Inter               |
| ZINTERCARD            | *TODO*              |
| ZDIFF                 | *TODO*              |
| ZRANGE                | Range               |
| ZRANGEBYSCORE         | IncrBy              |
| ZREVRANGEBYSCORE      | RevRangeByScore     |
| ZCOUNT                | Count               |
| ZREVRANGE             | RevRange            |
| ZCARD                 | Len                 |
| ZSCORE                | Score               |
| ZRANK                 | Rank                |
| ZREVRANK              | RevRank             |
| ZPOPMIN               | *TODO*              |
| ZPOPMAX               | *TODO*              |
| ZRANDMEMBER           | *TODO*              |

List of redis commands are generated from the following command:

```bash
cat redis/src/server.c | grep -o '"z.*",z.*Command' | grep -o '".*"' | cut -d '"' -f2
```

You may find that not all redis commands have corresponding go implementations,
the reason is as follows:

### Unsupported Commands

Redis' zset can operates elements in lexicographic order, which is not commonly
used function, so zset does not support commands like ZREMRANGEBYLEX, ZLEXCOUNT
and so on.

| Redis command         |
|-----------------------|
| ZREMRANGEBYLEX        |
| ZRANGEBYLEX           |
| ZREVRANGEBYLEX        |
| ZLEXCOUNT             |

In redis, user accesses zset via a string key. We do not need such string key
because we have variable. so the following commands are not implemented:

| Redis command         |
|-----------------------|
| ZUNIONSTORE           |
| ZINTERSTORE           |
| ZDIFFSTORE            |
| ZRANGESTORE           |
| ZMSCORE               |
| ZSCAN                 |

## QuickStart

```go
package main

import (
	"fmt"

	"github.com/bytedance/gopkg/collection/zset"
)

func main() {
	z := zset.NewFloat64()

	values := []string{"Alice", "Bob", "Zhang"}
	scores := []float64{90, 89, 59.9}
	for i := range values {
		z.Add(scores[i], values[i])
	}

	s, ok := z.Score("Alice")
	if ok {
		fmt.Println("Alice's score is", s)
	}

	n := z.Count(0, 60)
	fmt.Println("There are", n, "people below 60 points")

	for _, n := range z.Range(0, -1) {
		fmt.Println("zset range found", n.Value, n.Score)
	}
}
```
