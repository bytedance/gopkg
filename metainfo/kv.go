package metainfo

import "context"

type infoType int

const (
	invalidType           infoType = 0
	transientUpstreamType infoType = 1 << iota
	transientType
	persistentType
)

type ctxKeyType struct{}

var ctxKey ctxKeyType

type pair struct {
	pre  *pair
	mode infoType
	key  string
	val  string
}

func getKV(ctx context.Context) *pair {
	if ctx != nil {
		if kv, ok := ctx.Value(ctxKey).(*pair); ok {
			return kv
		}
	}
	return nil
}

func addKV(ctx context.Context, it infoType, key, val string) context.Context {
	if ctx == nil {
		return nil
	}
	return context.WithValue(ctx, ctxKey, &pair{
		pre:  getKV(ctx),
		mode: it,
		key:  key,
		val:  val,
	})
}

func getV(ctx context.Context, mode infoType, key string) (string, bool) {
	kv := getKV(ctx)
	for kv != nil {
		if kv.key == key && (kv.mode&mode) > 0 {
			return kv.val, len(kv.val) > 0
		}
		kv = kv.pre
	}
	return "", false
}

func getAll(ctx context.Context, mode infoType) map[string]string {
	kvs := make(map[string]string)
	del := make(map[string]struct{})
	kv := getKV(ctx)
	for kv != nil {
		if kv.mode&mode > 0 {
			if _, exist := kvs[kv.key]; !exist {
				kvs[kv.key] = kv.val
				if len(kv.val) == 0 {
					del[kv.key] = struct{}{}
				}
			}
		}
		kv = kv.pre
	}
	for k := range del {
		delete(kvs, k)
	}
	return kvs
}

// copyLink creates a new link copying the original one with each node processed by the given `modifier`.
// If modifier returns nil, the kv pair will be discard.
// Modifier must always return a valid pair object's pointer or nil.
func copyLink(ctx context.Context, modifier func(mode infoType, k, v string) *pair) context.Context {
	var n *pair
	pre := &n
	kv := getKV(ctx)
	for kv != nil {
		if p := modifier(kv.mode, kv.key, kv.val); p != nil {
			*pre = p
			pre = &p.pre
		}
		kv = kv.pre
	}
	// When n is nil after all, the whole link should be invisible in the new context
	return context.WithValue(ctx, ctxKey, n)
}
