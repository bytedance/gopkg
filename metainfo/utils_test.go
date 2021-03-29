package metainfo_test

import (
	"context"
	"testing"

	"github.com/bytedance/gopkg/metainfo"
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
