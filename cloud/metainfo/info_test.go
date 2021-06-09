// Copyright 2021 ByteDance Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metainfo_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

func TestWithValue(t *testing.T) {
	ctx := context.Background()

	k, v := "Key", "Value"
	ctx = metainfo.WithValue(ctx, k, v)
	assert(t, ctx != nil)

	x, ok := metainfo.GetValue(ctx, k)
	assert(t, ok)
	assert(t, x == v)
}

func TestWithEmpty(t *testing.T) {
	ctx := context.Background()

	k, v := "Key", "Value"
	ctx = metainfo.WithValue(ctx, k, "")
	assert(t, ctx != nil)

	_, ok := metainfo.GetValue(ctx, k)
	assert(t, !ok)

	ctx = metainfo.WithValue(ctx, "", v)
	assert(t, ctx != nil)

	_, ok = metainfo.GetValue(ctx, "")
	assert(t, !ok)
}

func TestDelValue(t *testing.T) {
	ctx := context.Background()

	k, v := "Key", "Value"
	ctx = metainfo.WithValue(ctx, k, v)
	assert(t, ctx != nil)

	x, ok := metainfo.GetValue(ctx, k)
	assert(t, ok)
	assert(t, x == v)

	ctx = metainfo.DelValue(ctx, k)
	assert(t, ctx != nil)

	x, ok = metainfo.GetValue(ctx, k)
	assert(t, !ok)

	assert(t, metainfo.DelValue(ctx, "") == ctx)
}

func TestGetAll(t *testing.T) {
	ctx := context.Background()

	ss := []string{"1", "2", "3"}
	for _, k := range ss {
		ctx = metainfo.WithValue(ctx, "key"+k, "val"+k)
	}

	m := metainfo.GetAllValues(ctx)
	assert(t, m != nil)
	assert(t, len(m) == len(ss))

	for _, k := range ss {
		assert(t, m["key"+k] == "val"+k)
	}
}

func TestGetAll2(t *testing.T) {
	ctx := context.Background()

	ss := []string{"1", "2", "3"}
	for _, k := range ss {
		ctx = metainfo.WithValue(ctx, "key"+k, "val"+k)
	}

	ctx = metainfo.DelValue(ctx, "key2")

	m := metainfo.GetAllValues(ctx)
	assert(t, m != nil)
	assert(t, len(m) == len(ss)-1)

	for _, k := range ss {
		if k == "2" {
			_, exist := m["key"+k]
			assert(t, !exist)
		} else {
			assert(t, m["key"+k] == "val"+k)
		}
	}
}

///////////////////////////////////////////////

func TestWithPersistentValue(t *testing.T) {
	ctx := context.Background()

	k, v := "Key", "Value"
	ctx = metainfo.WithPersistentValue(ctx, k, v)
	assert(t, ctx != nil)

	x, ok := metainfo.GetPersistentValue(ctx, k)
	assert(t, ok)
	assert(t, x == v)
}

func TestWithPersistentEmpty(t *testing.T) {
	ctx := context.Background()

	k, v := "Key", "Value"
	ctx = metainfo.WithPersistentValue(ctx, k, "")
	assert(t, ctx != nil)

	_, ok := metainfo.GetPersistentValue(ctx, k)
	assert(t, !ok)

	ctx = metainfo.WithPersistentValue(ctx, "", v)
	assert(t, ctx != nil)

	_, ok = metainfo.GetPersistentValue(ctx, "")
	assert(t, !ok)
}

func TestDelPersistentValue(t *testing.T) {
	ctx := context.Background()

	k, v := "Key", "Value"
	ctx = metainfo.WithPersistentValue(ctx, k, v)
	assert(t, ctx != nil)

	x, ok := metainfo.GetPersistentValue(ctx, k)
	assert(t, ok)
	assert(t, x == v)

	ctx = metainfo.DelPersistentValue(ctx, k)
	assert(t, ctx != nil)

	x, ok = metainfo.GetPersistentValue(ctx, k)
	assert(t, !ok)

	assert(t, metainfo.DelPersistentValue(ctx, "") == ctx)
}

func TestGetAllPersistent(t *testing.T) {
	ctx := context.Background()

	ss := []string{"1", "2", "3"}
	for _, k := range ss {
		ctx = metainfo.WithPersistentValue(ctx, "key"+k, "val"+k)
	}

	m := metainfo.GetAllPersistentValues(ctx)
	assert(t, m != nil)
	assert(t, len(m) == len(ss))

	for _, k := range ss {
		assert(t, m["key"+k] == "val"+k)
	}
}

func TestGetAllPersistent2(t *testing.T) {
	ctx := context.Background()

	ss := []string{"1", "2", "3"}
	for _, k := range ss {
		ctx = metainfo.WithPersistentValue(ctx, "key"+k, "val"+k)
	}

	ctx = metainfo.DelPersistentValue(ctx, "key2")

	m := metainfo.GetAllPersistentValues(ctx)
	assert(t, m != nil)
	assert(t, len(m) == len(ss)-1)

	for _, k := range ss {
		if k == "2" {
			_, exist := m["key"+k]
			assert(t, !exist)
		} else {
			assert(t, m["key"+k] == "val"+k)
		}
	}
}

///////////////////////////////////////////////

func TestNilSafty(t *testing.T) {
	assert(t, metainfo.TransferForward(nil) == nil)

	_, tOK := metainfo.GetValue(nil, "any")
	assert(t, !tOK)
	assert(t, metainfo.GetAllValues(nil) == nil)
	assert(t, metainfo.WithValue(nil, "any", "any") == nil)
	assert(t, metainfo.DelValue(nil, "any") == nil)

	_, pOK := metainfo.GetPersistentValue(nil, "any")
	assert(t, !pOK)
	assert(t, metainfo.GetAllPersistentValues(nil) == nil)
	assert(t, metainfo.WithPersistentValue(nil, "any", "any") == nil)
	assert(t, metainfo.DelPersistentValue(nil, "any") == nil)
}

func TestTransitAndPersistent(t *testing.T) {
	ctx := context.Background()

	ctx = metainfo.WithValue(ctx, "A", "a")
	ctx = metainfo.WithPersistentValue(ctx, "A", "b")

	x, xOK := metainfo.GetValue(ctx, "A")
	y, yOK := metainfo.GetPersistentValue(ctx, "A")

	assert(t, xOK)
	assert(t, yOK)
	assert(t, x == "a")
	assert(t, y == "b")

	_, uOK := metainfo.GetValue(ctx, "B")
	_, vOK := metainfo.GetPersistentValue(ctx, "B")

	assert(t, !uOK)
	assert(t, !vOK)

	ctx = metainfo.DelValue(ctx, "A")
	_, pOK := metainfo.GetValue(ctx, "A")
	q, qOK := metainfo.GetPersistentValue(ctx, "A")
	assert(t, !pOK)
	assert(t, qOK)
	assert(t, q == "b")
}

func TestTransferForward(t *testing.T) {
	ctx := context.Background()

	ctx = metainfo.WithValue(ctx, "A", "t")
	ctx = metainfo.WithPersistentValue(ctx, "A", "p")
	ctx = metainfo.WithValue(ctx, "A", "ta")
	ctx = metainfo.WithPersistentValue(ctx, "A", "pa")

	ctx = metainfo.TransferForward(ctx)
	assert(t, ctx != nil)

	x, xOK := metainfo.GetValue(ctx, "A")
	y, yOK := metainfo.GetPersistentValue(ctx, "A")

	assert(t, xOK)
	assert(t, yOK)
	assert(t, x == "ta")
	assert(t, y == "pa")

	ctx = metainfo.TransferForward(ctx)
	assert(t, ctx != nil)

	x, xOK = metainfo.GetValue(ctx, "A")
	y, yOK = metainfo.GetPersistentValue(ctx, "A")

	assert(t, !xOK)
	assert(t, yOK)
	assert(t, y == "pa")

	ctx = metainfo.WithValue(ctx, "B", "tb")

	ctx = metainfo.TransferForward(ctx)
	assert(t, ctx != nil)

	y, yOK = metainfo.GetPersistentValue(ctx, "A")
	z, zOK := metainfo.GetValue(ctx, "B")

	assert(t, yOK)
	assert(t, y == "pa")
	assert(t, zOK)
	assert(t, z == "tb")
}

///////////////////////////////////////////////

func initMetaInfo(count int) (context.Context, []string, []string) {
	ctx := context.Background()
	var keys, vals []string
	for i := 0; i < count; i++ {
		k, v := fmt.Sprintf("key-%d", i), fmt.Sprintf("val-%d", i)
		ctx = metainfo.WithValue(ctx, k, v)
		ctx = metainfo.WithPersistentValue(ctx, k, v)
		keys = append(keys, k)
		vals = append(vals, v)
	}
	return ctx, keys, vals
}

func benchmark(b *testing.B, api string, count int) {
	ctx, keys, vals := initMetaInfo(count)
	switch api {
	case "TransferForward":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = metainfo.TransferForward(ctx)
		}
	case "GetValue":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = metainfo.GetValue(ctx, keys[i%len(keys)])
		}
	case "GetAllValues":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = metainfo.GetAllValues(ctx)
		}
	case "WithValue":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = metainfo.WithValue(ctx, "key", "val")
		}
	case "WithValueAcc":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx = metainfo.WithValue(ctx, vals[i%len(vals)], "val")
		}
	case "DelValue":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = metainfo.DelValue(ctx, "key")
		}
	case "GetPersistentValue":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = metainfo.GetPersistentValue(ctx, keys[i%len(keys)])
		}
	case "GetAllPersistentValues":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = metainfo.GetAllPersistentValues(ctx)
		}
	case "WithPersistentValue":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = metainfo.WithPersistentValue(ctx, "key", "val")
		}
	case "WithPersistentValueAcc":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx = metainfo.WithPersistentValue(ctx, vals[i%len(vals)], "val")
		}
		_ = ctx
	case "DelPersistentValue":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = metainfo.DelPersistentValue(ctx, "key")
		}
	case "SaveMetaInfoToMap":
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m := make(map[string]string)
			metainfo.SaveMetaInfoToMap(ctx, m)
		}
	case "SetMetaInfoFromMap":
		m := make(map[string]string)
		c := context.Background()
		metainfo.SaveMetaInfoToMap(ctx, m)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = metainfo.SetMetaInfoFromMap(c, m)
		}
	}
}

func benchmarkParallel(b *testing.B, api string, count int) {
	ctx, keys, vals := initMetaInfo(count)
	switch api {
	case "TransferForward":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = metainfo.TransferForward(ctx)
			}
		})
	case "GetValue":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var i int
			for pb.Next() {
				_, _ = metainfo.GetValue(ctx, keys[i%len(keys)])
				i++
			}
		})
	case "GetAllValues":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = metainfo.GetAllValues(ctx)
			}
		})
	case "WithValue":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = metainfo.WithValue(ctx, "key", "val")
			}
		})
	case "WithValueAcc":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			tmp := ctx
			var i int
			for pb.Next() {
				tmp = metainfo.WithValue(tmp, vals[i%len(vals)], "val")
				i++
			}
		})
	case "DelValue":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = metainfo.DelValue(ctx, "key")
			}
		})
	case "GetPersistentValue":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			var i int
			for pb.Next() {
				_, _ = metainfo.GetPersistentValue(ctx, keys[i%len(keys)])
				i++
			}
		})
	case "GetAllPersistentValues":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = metainfo.GetAllPersistentValues(ctx)
			}
		})
	case "WithPersistentValue":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = metainfo.WithPersistentValue(ctx, "key", "val")
			}
		})
	case "WithPersistentValueAcc":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			tmp := ctx
			var i int
			for pb.Next() {
				tmp = metainfo.WithPersistentValue(tmp, vals[i%len(vals)], "val")
				i++
			}
		})
	case "DelPersistentValue":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = metainfo.DelPersistentValue(ctx, "key")
			}
		})
	case "SaveMetaInfoToMap":
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m := make(map[string]string)
				metainfo.SaveMetaInfoToMap(ctx, m)
			}
		})
	case "SetMetaInfoFromMap":
		m := make(map[string]string)
		c := context.Background()
		metainfo.SaveMetaInfoToMap(ctx, m)
		b.ReportAllocs()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = metainfo.SetMetaInfoFromMap(c, m)
			}
		})
	}
}

func BenchmarkAll(b *testing.B) {
	APIs := []string{
		"TransferForward",
		"GetValue",
		"GetAllValues",
		"WithValue",
		"WithValueAcc",
		"DelValue",
		"GetPersistentValue",
		"GetAllPersistentValues",
		"WithPersistentValue",
		"WithPersistentValueAcc",
		"DelPersistentValue",
		"SaveMetaInfoToMap",
		"SetMetaInfoFromMap",
	}
	for _, api := range APIs {
		for _, cnt := range []int{10, 20, 50, 100} {
			fun := fmt.Sprintf("%s_%d", api, cnt)
			b.Run(fun, func(b *testing.B) { benchmark(b, api, cnt) })
		}
	}
}

func BenchmarkAllParallel(b *testing.B) {
	APIs := []string{
		"TransferForward",
		"GetValue",
		"GetAllValues",
		"WithValue",
		"WithValueAcc",
		"DelValue",
		"GetPersistentValue",
		"GetAllPersistentValues",
		"WithPersistentValue",
		"WithPersistentValueAcc",
		"DelPersistentValue",
		"SaveMetaInfoToMap",
		"SetMetaInfoFromMap",
	}
	for _, api := range APIs {
		for _, cnt := range []int{10, 20, 50, 100} {
			fun := fmt.Sprintf("%s_%d", api, cnt)
			b.Run(fun, func(b *testing.B) { benchmarkParallel(b, api, cnt) })
		}
	}
}
