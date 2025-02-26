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
	"net/http"
	"testing"

	"github.com/bytedance/gopkg/cloud/metainfo"
)

func TestHTTPHeaderToCGIVariable(t *testing.T) {
	for k, v := range map[string]string{
		"a":           "A",
		"aBc":         "ABC",
		"a1z":         "A1Z",
		"ab-":         "AB_",
		"-cd":         "_CD",
		"abc-def":     "ABC_DEF",
		"Abc-def_ghi": "ABC_DEF_GHI",
	} {
		assert(t, metainfo.HTTPHeaderToCGIVariable(k) == v)
	}
}

func TestCGIVariableToHTTPHeader(t *testing.T) {
	for k, v := range map[string]string{
		"a":           "a",
		"aBc":         "abc",
		"a1z":         "a1z",
		"AB_":         "ab-",
		"_CD":         "-cd",
		"ABC_DEF":     "abc-def",
		"ABC-def_GHI": "abc-def-ghi",
	} {
		assert(t, metainfo.CGIVariableToHTTPHeader(k) == v)
	}
}

func TestFromHTTPHeader(t *testing.T) {
	assert(t, metainfo.FromHTTPHeader(nil, nil) == nil)

	h := make(http.Header)
	c := context.Background()
	c1 := metainfo.FromHTTPHeader(c, metainfo.HTTPHeader(h))
	assert(t, c == c1)

	h.Set("abc", "def")
	h.Set(metainfo.HTTPPrefixTransient+"123", "456")
	h.Set(metainfo.HTTPPrefixTransient+"abc-def", "ghi")
	h.Set(metainfo.HTTPPrefixPersistent+"xyz", "000")
	c1 = metainfo.FromHTTPHeader(c, metainfo.HTTPHeader(h))
	assert(t, c != c1)
	vs := metainfo.GetAllValues(c1)
	assert(t, len(vs) == 2, vs)
	assert(t, vs["ABC_DEF"] == "ghi" && vs["123"] == "456", vs)
	vs = metainfo.GetAllPersistentValues(c1)
	assert(t, len(vs) == 1 && vs["XYZ"] == "000")
}

func TestFromHTTPHeaderKeepPreviousData(t *testing.T) {
	c0 := context.Background()
	c0 = metainfo.WithValue(c0, "uk", "uv")
	c0 = metainfo.TransferForward(c0)
	c0 = metainfo.WithValue(c0, "tk", "tv")
	c0 = metainfo.WithPersistentValue(c0, "pk", "pv")

	h := make(http.Header)
	h.Set(metainfo.HTTPPrefixTransient+"xk", "xv")
	h.Set(metainfo.HTTPPrefixPersistent+"yk", "yv")
	h.Set(metainfo.HTTPPrefixPersistent+"pk", "pp")

	c1 := metainfo.FromHTTPHeader(c0, metainfo.HTTPHeader(h))
	assert(t, c0 != c1)
	vs := metainfo.GetAllValues(c1)
	assert(t, len(vs) == 3, len(vs))
	assert(t, vs["tk"] == "tv" && vs["uk"] == "uv" && vs["XK"] == "xv")
	vs = metainfo.GetAllPersistentValues(c1)
	assert(t, len(vs) == 3)
	assert(t, vs["pk"] == "pv" && vs["YK"] == "yv" && vs["PK"] == "pp")
}

func TestToHTTPHeader(t *testing.T) {
	metainfo.ToHTTPHeader(nil, nil)

	h := make(http.Header)
	c := context.Background()
	metainfo.ToHTTPHeader(c, h)
	assert(t, len(h) == 0)

	c = metainfo.WithValue(c, "123", "456")
	c = metainfo.WithPersistentValue(c, "abc", "def")
	metainfo.ToHTTPHeader(c, h)
	assert(t, len(h) == 2)
	assert(t, h.Get(metainfo.HTTPPrefixTransient+"123") == "456")
	assert(t, h.Get(metainfo.HTTPPrefixPersistent+"abc") == "def")
}

func TestHTTPHeader(t *testing.T) {
	h := make(metainfo.HTTPHeader)
	h.Set("Hello", "halo")
	h.Set("hello", "world")

	kvs := make(map[string]string)
	h.Visit(func(k, v string) {
		kvs[k] = v
	})
	assert(t, len(kvs) == 1)
	assert(t, kvs["hello"] == "world")
}

func BenchmarkHTTPHeaderToCGIVariable(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metainfo.HTTPHeaderToCGIVariable(metainfo.HTTPPrefixPersistent + "hello-world")
	}
}

func BenchmarkCGIVariableToHTTPHeader(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = metainfo.CGIVariableToHTTPHeader(metainfo.PrefixPersistent + "HELLO_WORLD")
	}
}

func BenchmarkFromHTTPHeader(b *testing.B) {
	for _, cnt := range []int{0, 10, 20, 50, 100} {
		hd := make(metainfo.HTTPHeader)
		hd.Set("content-type", "test")
		hd.Set("content-length", "12345")
		for i := 0; len(hd) < cnt; i++ {
			hd.Set(metainfo.HTTPPrefixTransient+fmt.Sprintf("tk%d", i), fmt.Sprintf("tv-%d", i))
			hd.Set(metainfo.HTTPPrefixPersistent+fmt.Sprintf("pk%d", i), fmt.Sprintf("pv-%d", i))
		}
		ctx := context.Background()
		fun := fmt.Sprintf("FromHTTPHeader-%d", cnt)
		b.Run(fun, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = metainfo.FromHTTPHeader(ctx, hd)
			}
		})
	}
}

func BenchmarkToHTTPHeader(b *testing.B) {
	for _, cnt := range []int{0, 10, 20, 50, 100} {
		ctx, _, _ := initMetaInfo(cnt)
		fun := fmt.Sprintf("ToHTTPHeader-%d", cnt)
		b.Run(fun, func(b *testing.B) {
			hd := make(metainfo.HTTPHeader)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				metainfo.ToHTTPHeader(ctx, hd)
			}
		})
	}
}
