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

func TestHasMetaInfo(t *testing.T) {
	c0 := context.Background()
	assert(t, !metainfo.HasMetaInfo(c0))

	c1 := metainfo.WithValue(c0, "k", "v")
	assert(t, metainfo.HasMetaInfo(c1))

	c2 := metainfo.WithPersistentValue(c0, "k", "v")
	assert(t, metainfo.HasMetaInfo(c2))
}

func TestSetMetaInfoFromMap(t *testing.T) {
	// Nil tests
	assert(t, metainfo.SetMetaInfoFromMap(nil, nil) == nil)

	ctx := context.Background()
	assert(t, metainfo.SetMetaInfoFromMap(ctx, nil) == ctx)

	// Ignore ill-format keys
	m := map[string]string{
		"foo-key": "foo-val",
		"bar-key": "bar-val",
	}
	assert(t, metainfo.SetMetaInfoFromMap(ctx, m) == ctx)

	// Ignore empty keys
	m[metainfo.PrefixTransientUpstream] = "1"
	m[metainfo.PrefixTransient] = "2"
	m[metainfo.PrefixPersistent] = "3"
	assert(t, metainfo.SetMetaInfoFromMap(ctx, m) == ctx)

	// Ignore empty values
	k1 := metainfo.PrefixTransientUpstream + "k1"
	k2 := metainfo.PrefixTransient + "k2"
	k3 := metainfo.PrefixPersistent + "k3"
	m[k1] = ""
	m[k2] = ""
	m[k3] = ""
	assert(t, metainfo.SetMetaInfoFromMap(ctx, m) == ctx)

	// Accept valid key-value pairs
	k4 := metainfo.PrefixTransientUpstream + "k4"
	k5 := metainfo.PrefixTransient + "k5"
	k6 := metainfo.PrefixPersistent + "k6"
	m[k4] = "v4"
	m[k5] = "v5"
	m[k6] = "v6"
	ctx2 := metainfo.SetMetaInfoFromMap(ctx, m)
	assert(t, ctx2 != ctx)

	ctx = ctx2

	v1, ok1 := metainfo.GetValue(ctx, "k4")
	v2, ok2 := metainfo.GetValue(ctx, "k5")
	_, ok3 := metainfo.GetValue(ctx, "k6")
	assert(t, ok1)
	assert(t, ok2)
	assert(t, !ok3)
	assert(t, v1 == "v4")
	assert(t, v2 == "v5")

	_, ok4 := metainfo.GetPersistentValue(ctx, "k4")
	_, ok5 := metainfo.GetPersistentValue(ctx, "k5")
	v3, ok6 := metainfo.GetPersistentValue(ctx, "k6")
	assert(t, !ok4)
	assert(t, !ok5)
	assert(t, ok6)
	assert(t, v3 == "v6")
}

func TestSetMetaInfoFromMapKeepPreviousData(t *testing.T) {
	ctx0 := context.Background()
	ctx0 = metainfo.WithValue(ctx0, "uk", "uv")
	ctx0 = metainfo.TransferForward(ctx0)
	ctx0 = metainfo.WithValue(ctx0, "tk", "tv")
	ctx0 = metainfo.WithPersistentValue(ctx0, "pk", "pv")

	m := map[string]string{
		metainfo.PrefixTransientUpstream + "xk": "xv",
		metainfo.PrefixTransient + "yk":         "yv",
		metainfo.PrefixPersistent + "zk":        "zv",
		metainfo.PrefixTransient + "uk":         "vv", // overwrite "uk"
	}
	ctx1 := metainfo.SetMetaInfoFromMap(ctx0, m)
	assert(t, ctx1 != ctx0)

	ts := metainfo.GetAllValues(ctx1)
	ps := metainfo.GetAllPersistentValues(ctx1)
	assert(t, len(ts) == 4)
	assert(t, len(ps) == 2)

	assert(t, ts["uk"] == "vv")
	assert(t, ts["tk"] == "tv")
	assert(t, ts["xk"] == "xv")
	assert(t, ts["yk"] == "yv")
	assert(t, ps["pk"] == "pv")
	assert(t, ps["zk"] == "zv")
}

func TestSaveMetaInfoToMap(t *testing.T) {
	m := make(map[string]string)

	metainfo.SaveMetaInfoToMap(nil, m)
	assert(t, len(m) == 0)

	ctx := context.Background()
	metainfo.SaveMetaInfoToMap(ctx, m)
	assert(t, len(m) == 0)

	ctx = metainfo.WithValue(ctx, "a", "a")
	ctx = metainfo.WithValue(ctx, "b", "b")
	ctx = metainfo.WithValue(ctx, "a", "a2")
	ctx = metainfo.WithValue(ctx, "b", "b2")
	ctx = metainfo.WithPersistentValue(ctx, "a", "a")
	ctx = metainfo.WithPersistentValue(ctx, "b", "b")
	ctx = metainfo.WithPersistentValue(ctx, "a", "a3")
	ctx = metainfo.WithPersistentValue(ctx, "b", "b3")
	ctx = metainfo.DelValue(ctx, "a")
	ctx = metainfo.DelPersistentValue(ctx, "a")

	metainfo.SaveMetaInfoToMap(ctx, m)
	assert(t, len(m) == 2)
	assert(t, m[metainfo.PrefixTransient+"b"] == "b2")
	assert(t, m[metainfo.PrefixPersistent+"b"] == "b3")
}

func BenchmarkSetMetaInfoFromMap(b *testing.B) {
	ctx := metainfo.WithPersistentValue(context.Background(), "key", "val")
	m := map[string]string{}
	for i := 0; i < 32; i++ {
		m[fmt.Sprintf("key-%d", i)] = fmt.Sprintf("val-%d", i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metainfo.SetMetaInfoFromMap(ctx, m)
	}
}
