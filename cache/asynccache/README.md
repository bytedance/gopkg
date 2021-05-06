# asynccache

## Introduction

`asynccache` fetches and updates the latest data periodically and supports expire a key if unused for a period.

The functions it provides is listed below:
```go
type AsyncCache interface {
	// SetDefault sets the default value of given key if it is new to the cache.
	// It is useful for cache warming up.
	// Param val should not be nil.
	SetDefault(key string, val interface{}) (exist bool)

	// Get tries to fetch a value corresponding to the given key from the cache.
	// If error occurs during the first time fetching, it will be cached until the
	// sequential fetching triggered by the refresh goroutine succeed.
	Get(key string) (val interface{}, err error)

	// GetOrSet tries to fetch a value corresponding to the given key from the cache.
	// If the key is not yet cached or error occurs, the default value will be set.
	GetOrSet(key string, defaultVal interface{}) (val interface{})

	// Dump dumps all cache entries.
	// This will not cause expire to refresh.
	Dump() map[string]interface{}

	// DeleteIf deletes cached entries that match the `shouldDelete` predicate.
	DeleteIf(shouldDelete func(key string) bool)

	// Close closes the async cache.
	// This should be called when the cache is no longer needed, or may lead to resource leak.
	Close()
}
```

## Example

```go
var key, ret = "key", "ret"
opt := Options{
    RefreshDuration: time.Second,
    IsSame: func(key string, oldData, newData interface{}) bool {
        return false
    },
    Fetcher: func(key string) (interface{}, error) {
        return ret, nil
    },
}
c := NewAsyncCache(opt)

v, err := c.Get(key)
assert.NoError(err)
assert.Equal(v.(string), ret)

time.Sleep(time.Second / 2)
ret = "change"
v, err = c.Get(key)
assert.NoError(err)
assert.NotEqual(v.(string), ret)

time.Sleep(time.Second)
v, err = c.Get(key)
assert.NoError(err)
assert.Equal(v.(string), ret)
```
