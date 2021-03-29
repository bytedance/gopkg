package metainfo

import (
	"context"
)

// The prefix listed below may be used to tag the types of values when there is no context to carry them.
const (
	PrefixPersistent        = "RPC_PERSIST_"
	PrefixTransient         = "RPC_TRANSIT_"
	PrefixTransientUpstream = "RPC_TRANSIT_UPSTREAM_"
)

// **Using empty string as key or value is not support.**

// TransferForward converts transient values to transient-upstream values and filters out original transient-upstream values.
// It should be used before the context is passing from server to client.
func TransferForward(ctx context.Context) context.Context {
	if ctx == nil {
		return nil
	}
	return copyLink(ctx, func(mode infoType, k, v string) *pair {
		if mode == transientUpstreamType {
			return nil
		}
		if mode == transientType {
			mode = transientUpstreamType
		}
		return &pair{
			mode: mode,
			key:  k,
			val:  v,
		}
	})
}

// GetValue retrieves the value set into the context by the given key.
func GetValue(ctx context.Context, k string) (string, bool) {
	return getV(ctx, transientUpstreamType|transientType, k)
}

// GetAllValues retrieves all transient values
func GetAllValues(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	return getAll(ctx, transientUpstreamType|transientType)
}

// WithValue sets the value into the context by the given key.
// This value will be propagated to the next service/endpoint through an RPC call.
//
// Notice that it will not propagate any further beyond the next service/endpoint,
// Use WithPersistentValue if you want to pass a key/value pair all the way.
func WithValue(ctx context.Context, k, v string) context.Context {
	if len(k) == 0 || len(v) == 0 {
		return ctx
	}
	return addKV(ctx, transientType, k, v)
}

// DelValue deletes a key/value from the current context.
// Since empty string value is not valid, we could just set the value to be empty.
func DelValue(ctx context.Context, k string) context.Context {
	if len(k) == 0 {
		return ctx
	}
	return addKV(ctx, transientType, k, "")
}

// GetPersistentValue retrieves the persistent value set into the context by the given key.
func GetPersistentValue(ctx context.Context, k string) (string, bool) {
	return getV(ctx, persistentType, k)
}

// GetAllPersistentValues retrieves all persistent values.
func GetAllPersistentValues(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	return getAll(ctx, persistentType)
}

// WithPersistentValue sets the value info the context by the given key.
// This value will be propagated to the services along the RPC call chain.
func WithPersistentValue(ctx context.Context, k, v string) context.Context {
	if len(k) == 0 || len(v) == 0 {
		return ctx
	}
	return addKV(ctx, persistentType, k, v)
}

// DelPersistentValue deletes a persistent key/value from the current context.
// Since empty string value is not valid, we could just set the value to be empty.
func DelPersistentValue(ctx context.Context, k string) context.Context {
	if len(k) == 0 {
		return ctx
	}
	return addKV(ctx, persistentType, k, "")
}
