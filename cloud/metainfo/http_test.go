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
