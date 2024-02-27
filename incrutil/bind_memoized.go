package incrutil

import (
	"context"
	"sync"

	"github.com/wcharczuk/go-incr"
)

// BindMemoized returns a node that caches the results of the bind function, and as a result the input must be typed such that the
// computed value of the input is comparable.
func BindMemoized[A comparable, B any](scope incr.Scope, a incr.Incr[A], fn incr.BindFunc[A, B]) BindMemoizedIncr[A, B] {
	return BindMemoizedContextCached[A, B](scope, a, func(_ context.Context, innerScope incr.Scope, av A) (incr.Incr[B], error) {
		return fn(innerScope, av), nil
	}, new(mapCache[A, incr.Incr[B]]))
}

// BindMemoizedCached returns a node that caches the results of the bind function, and as a result the input must be typed such that the
// computed value of the input is comparable. The provided cache reference will be used, allowing this node to share its cache with
// other bind nodes.
func BindMemoizedCached[A comparable, B any](scope incr.Scope, a incr.Incr[A], fn incr.BindFunc[A, B], cache BindCache[A, B]) BindMemoizedIncr[A, B] {
	return BindMemoizedContextCached[A, B](scope, a, func(_ context.Context, innerScope incr.Scope, av A) (incr.Incr[B], error) {
		return fn(innerScope, av), nil
	}, cache)
}

// BindMemoizedContext returns a node that caches the results of the bind function, which takes a context and returns an error, and as a
// result the input must be typed such that the computed value of the input is comparable.
func BindMemoizedContext[A comparable, B any](scope incr.Scope, a incr.Incr[A], fn incr.BindContextFunc[A, B]) BindMemoizedIncr[A, B] {
	return BindMemoizedContextCached[A, B](scope, a, func(ctx context.Context, innerScope incr.Scope, av A) (incr.Incr[B], error) {
		return fn(ctx, innerScope, av)
	}, new(mapCache[A, incr.Incr[B]]))
}

// BindMemoizedContextCached returns a memoized bind node.
func BindMemoizedContextCached[A comparable, B any](scope incr.Scope, a incr.Incr[A], fn incr.BindContextFunc[A, B], cache BindCache[A, B]) BindMemoizedIncr[A, B] {
	bm := new(bindMemoizedIncr[A, B])
	bm.cache = cache
	bm.BindIncr = incr.BindContext(scope, a, func(ctx context.Context, innerScope incr.Scope, key A) (incr.Incr[B], error) {
		if cached, ok := cache.Get(key); ok {
			return cached, nil
		}
		value, err := fn(ctx, innerScope, key)
		if err != nil {
			return nil, err
		}
		cache.Put(key, value)
		return value, nil
	})
	return bm
}

// BindCache is a type that can implement a cache for `BindMemoized`.
type BindCache[A comparable, B any] interface {
	Get(A) (incr.Incr[B], bool)
	Put(A, incr.Incr[B])
	Purge(A)
	Clear()
}

var (
	_ incr.IParents = (*bindMemoizedIncr[int, any])(nil)
)

// BindMemoizedIncr is a type that walks like a bind and quacks like a bind
// but actually implements caching under the hood for returned bind nodes.
type BindMemoizedIncr[A comparable, B any] interface {
	incr.BindIncr[B]
	Purge(A)
	Clear()
}

type bindMemoizedIncr[A comparable, B any] struct {
	incr.BindIncr[B]
	cache BindCache[A, B]
}

func (bmi *bindMemoizedIncr[A, B]) Purge(k A) {
	bmi.cache.Purge(k)
}

func (bmi *bindMemoizedIncr[A, B]) Clear() {
	bmi.cache.Clear()
}

// mapCache is a map backed cache that is _incredibly_ basic.
type mapCache[A comparable, B any] struct {
	cache   map[A]B
	cacheMu sync.Mutex
}

func (mc *mapCache[A, B]) Get(key A) (value B, ok bool) {
	mc.cacheMu.Lock()
	defer mc.cacheMu.Unlock()
	if mc.cache == nil {
		return
	}
	value, ok = mc.cache[key]
	return
}

func (mc *mapCache[A, B]) Put(key A, value B) {
	mc.cacheMu.Lock()
	defer mc.cacheMu.Unlock()
	if mc.cache == nil {
		mc.cache = make(map[A]B)
	}
	mc.cache[key] = value
}

func (mc *mapCache[A, B]) Purge(key A) {
	mc.cacheMu.Lock()
	defer mc.cacheMu.Unlock()
	if mc.cache == nil {
		return
	}
	delete(mc.cache, key)
}

func (mc *mapCache[A, B]) Clear() {
	mc.cacheMu.Lock()
	defer mc.cacheMu.Unlock()
	if mc.cache == nil {
		return
	}
	clear(mc.cache)
}
