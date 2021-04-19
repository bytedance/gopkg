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

package asynccache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetOK(t *testing.T) {
	var key, ret = "key", "ret"
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return ret, nil
		},
	}
	c := NewAsyncCache(op)

	v, err := c.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, v.(string), ret)

	time.Sleep(time.Second / 2)
	ret = "change"
	v, err = c.Get(key)
	assert.NoError(t, err)
	assert.NotEqual(t, v.(string), ret)

	time.Sleep(time.Second)
	v, err = c.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, v.(string), ret)
}

func TestGetErr(t *testing.T) {
	var key, ret = "key", "ret"
	var first = true
	op := Options{
		RefreshDuration: time.Second + 100*time.Millisecond,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			if first {
				first = false
				return nil, errors.New("error")
			}
			return ret, nil
		},
	}
	c := NewAsyncCache(op)

	v, err := c.Get(key)
	assert.Error(t, err)
	assert.Nil(t, v)

	time.Sleep(time.Second / 2)
	_, err2 := c.Get(key)
	assert.Equal(t, err, err2)

	time.Sleep(time.Second + 10*time.Millisecond)
	v, err = c.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, v.(string), ret)
}

func TestGetOrSetOK(t *testing.T) {
	var key, ret, def = "key", "ret", "def"
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return ret, nil
		},
	}
	c := NewAsyncCache(op)

	v := c.GetOrSet(key, def)
	assert.Equal(t, v.(string), ret)

	time.Sleep(time.Second / 2)
	ret = "change"
	v = c.GetOrSet(key, def)
	assert.NotEqual(t, v.(string), ret)

	time.Sleep(time.Second)
	v = c.GetOrSet(key, def)
	assert.Equal(t, v.(string), ret)
}

func TestGetOrSetErr(t *testing.T) {
	var key, ret, def = "key", "ret", "def"
	var first = true
	op := Options{
		RefreshDuration: time.Second + 500*time.Millisecond,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			if first {
				first = false
				return nil, errors.New("error")
			}
			return ret, nil
		},
	}
	c := NewAsyncCache(op)

	v := c.GetOrSet(key, def)
	assert.Equal(t, v.(string), def)

	time.Sleep(time.Second / 2)
	v = c.GetOrSet(key, ret)
	assert.NotEqual(t, v.(string), ret)
	assert.Equal(t, v.(string), def)

	time.Sleep(time.Second + 500*time.Millisecond)
	v = c.GetOrSet(key, def)
	assert.Equal(t, v.(string), ret)
}

func TestSetDefault(t *testing.T) {
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return nil, errors.New("error")
		},
	}
	c := NewAsyncCache(op)

	v := c.GetOrSet("key1", "def1")
	assert.Equal(t, v.(string), "def1")

	exist := c.SetDefault("key2", "val2")
	assert.False(t, exist)
	v = c.GetOrSet("key2", "def2")
	assert.Equal(t, v.(string), "val2")

	exist = c.SetDefault("key2", "val3")
	assert.True(t, exist)
	v = c.GetOrSet("key2", "def2")
	assert.Equal(t, v.(string), "val2")
}

func TestDeleteIf(t *testing.T) {
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return nil, errors.New("error")
		},
	}
	c := NewAsyncCache(op)

	c.SetDefault("key", "val")
	v := c.GetOrSet("key", "def")
	assert.Equal(t, v.(string), "val")

	d, _ := c.(interface{ DeleteIf(func(key string) bool) })
	d.DeleteIf(func(string) bool { return true })

	v = c.GetOrSet("key", "def")
	assert.Equal(t, v.(string), "def")
}

func TestClose(t *testing.T) {
	var dur = time.Second / 10
	var cnt int
	op := Options{
		RefreshDuration: dur - 10*time.Millisecond,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			cnt++
			return cnt, nil
		},
		EnableExpire:   true,
		ExpireDuration: time.Second,
	}
	c := NewAsyncCache(op)

	v := c.GetOrSet("key", 10)
	assert.Equal(t, v.(int), 1)

	time.Sleep(dur)
	v = c.GetOrSet("key", 10)
	assert.Equal(t, v.(int), 2)

	time.Sleep(dur)
	v = c.GetOrSet("key", 10)
	assert.Equal(t, v.(int), 3)

	c.Close()

	time.Sleep(dur)
	v = c.GetOrSet("key", 10)
	assert.Equal(t, v.(int), 3)
}

func TestExpire(t *testing.T) {
	// trigger is used to mark whether fetcher is called
	trigger := false
	op := Options{
		EnableExpire:    true,
		ExpireDuration:  3 * time.Minute,
		RefreshDuration: time.Minute,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return true
		},
		Fetcher: func(key string) (interface{}, error) {
			trigger = true
			return "", nil
		},
	}
	c := NewAsyncCache(op).(*asyncCache)

	// GetOrSet cannot trigger fetcher when SetDefault before
	c.SetDefault("key-default", "")
	c.SetDefault("key-alive", "")
	c.GetOrSet("key-alive", "")
	assert.False(t, trigger)

	c.Get("key-expire")
	assert.True(t, trigger)

	// first expire set tag
	c.expire()

	trigger = false
	c.Get("key-alive")
	assert.False(t, trigger)
	// second expire, both key-default & key-expire have been removed
	c.expire()
	c.refresh() // prove refresh does not affect expire

	trigger = false
	c.Get("key-alive")
	assert.False(t, trigger)
	trigger = false
	c.Get("key-default")
	assert.True(t, trigger)
	trigger = false
	c.Get("key-expire")
	assert.True(t, trigger)
}

func BenchmarkGet(b *testing.B) {
	var key = "key"
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return "", nil
		},
	}
	c := NewAsyncCache(op)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Get(key)
	}
}

func BenchmarkGetParallel(b *testing.B) {
	var key = "key"
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return "", nil
		},
	}
	c := NewAsyncCache(op)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = c.Get(key)
		}
	})
}

func BenchmarkGetOrSet(b *testing.B) {
	var key, def = "key", "def"
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return "", nil
		},
	}
	c := NewAsyncCache(op)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.GetOrSet(key, def)
	}
}

func BenchmarkGetOrSetParallel(b *testing.B) {
	var key, def = "key", "def"
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return "", nil
		},
	}
	c := NewAsyncCache(op)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = c.GetOrSet(key, def)
		}
	})
}

func BenchmarkRefresh(b *testing.B) {
	var key, def = "key", "def"
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return "", nil
		},
	}
	c := NewAsyncCache(op).(*asyncCache)
	c.SetDefault(key, def)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.refresh()
	}
}

func BenchmarkRefreshParallel(b *testing.B) {
	var key, def = "key", "def"
	op := Options{
		RefreshDuration: time.Second,
		IsSame: func(key string, oldData, newData interface{}) bool {
			return false
		},
		Fetcher: func(key string) (interface{}, error) {
			return "", nil
		},
	}
	c := NewAsyncCache(op).(*asyncCache)
	c.SetDefault(key, def)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.refresh()
		}
	})
}
