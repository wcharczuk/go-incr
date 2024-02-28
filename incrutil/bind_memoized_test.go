package incrutil

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_BindMemoized(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	bv := incr.Var(g, "a")

	var called uint32
	bm := BindMemoized(g, bv, func(bs incr.Scope, av string) incr.Incr[string] {
		atomic.AddUint32(&called, 1)
		if av == "a" {
			return incr.Return(bs, "a-value")
		}
		return incr.Return(bs, "other-value")
	})
	obm := incr.MustObserve(g, bm)

	// shouldn't do anything
	bm.Cache().Purge("doesn't exist")
	// shouldn't do anything
	bm.Cache().Clear()

	testutil.Equal(t, 0, called)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 1, called)
	testutil.Equal(t, "a-value", obm.Value())

	bv.Set("b")
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 2, called)
	testutil.Equal(t, "other-value", obm.Value())

	bv.Set("a")
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 2, called)
	testutil.Equal(t, "a-value", obm.Value())

	bv.Set("a") // we have to trigger staleness
	bm.Cache().Purge("a")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 3, called)
	testutil.Equal(t, "a-value", obm.Value())

	bv.Set("b")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 3, called)
	testutil.Equal(t, "other-value", obm.Value())

	bv.Set("b")
	bm.Cache().Clear()

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 4, called)
	testutil.Equal(t, "other-value", obm.Value())

	bv.Set("a")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 5, called)
	testutil.Equal(t, "a-value", obm.Value())

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 5, called)
	testutil.Equal(t, "a-value", obm.Value())
}

func Test_BindMemoizedCached(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	cache := new(mapCache[string, incr.Incr[string]])
	bv := incr.Var(g, "a")

	var called uint32
	bm := BindMemoizedCached(g, bv, func(bs incr.Scope, av string) incr.Incr[string] {
		atomic.AddUint32(&called, 1)
		if av == "a" {
			return incr.Return(bs, "a-value")
		}
		return incr.Return(bs, "other-value")
	}, cache)
	obm := incr.MustObserve(g, bm)

	testutil.Equal(t, 0, called)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 1, called)
	testutil.Equal(t, "a-value", obm.Value())

	testutil.Equal(t, 1, len(cache.cache))

	bv.Set("b")
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 2, called)
	testutil.Equal(t, "other-value", obm.Value())

	testutil.Equal(t, 2, len(cache.cache))

	bv.Set("c")
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 3, called)
	testutil.Equal(t, "other-value", obm.Value())

	testutil.Equal(t, 3, len(cache.cache))
}

func Test_BindMemoizedContext(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	bv := incr.Var(g, "a")
	var called uint32
	bm := BindMemoizedContext(g, bv, func(ctx context.Context, bs incr.Scope, av string) (incr.Incr[string], error) {
		testutil.BlueDye(ctx, t)
		atomic.AddUint32(&called, 1)
		if av == "a" {
			return incr.Return(bs, "a-value"), nil
		}
		return incr.Return(bs, "other-value"), nil
	})
	obm := incr.MustObserve(g, bm)

	testutil.Equal(t, 0, called)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 1, called)
	testutil.Equal(t, "a-value", obm.Value())

	bv.Set("b")
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 2, called)
	testutil.Equal(t, "other-value", obm.Value())

	bv.Set("c")
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 3, called)
	testutil.Equal(t, "other-value", obm.Value())
}

func Test_BindMemoizedContextCached(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	cache := new(mapCache[string, incr.Incr[string]])
	bv := incr.Var(g, "a")

	var called uint32
	bm := BindMemoizedContextCached(g, bv, func(ctx context.Context, bs incr.Scope, av string) (incr.Incr[string], error) {
		testutil.BlueDye(ctx, t)
		atomic.AddUint32(&called, 1)
		if av == "a" {
			return incr.Return(bs, "a-value"), nil
		}
		return incr.Return(bs, "other-value"), nil
	}, cache)
	obm := incr.MustObserve(g, bm)

	testutil.Equal(t, 0, called)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 1, called)
	testutil.Equal(t, "a-value", obm.Value())

	testutil.Equal(t, 1, len(cache.cache))

	bv.Set("b")
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 2, called)
	testutil.Equal(t, "other-value", obm.Value())

	testutil.Equal(t, 2, len(cache.cache))

	bv.Set("c")
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 3, called)
	testutil.Equal(t, "other-value", obm.Value())

	testutil.Equal(t, 3, len(cache.cache))
}

func Test_BindMemoizedContext_error(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	cache := new(mapCache[string, incr.Incr[string]])

	bv := incr.Var(g, "a")
	var called uint32
	bm := BindMemoizedContextCached(g, bv, func(ctx context.Context, bs incr.Scope, av string) (incr.Incr[string], error) {
		atomic.AddUint32(&called, 1)
		return nil, fmt.Errorf("nil")
	}, cache)
	obm := incr.MustObserve(g, bm)

	testutil.Equal(t, 0, called)

	err := g.Stabilize(ctx)
	testutil.Error(t, err)
	testutil.Equal(t, 1, called)
	testutil.Equal(t, "", obm.Value())
	testutil.Equal(t, 0, len(cache.cache))

	bv.Set("b")
	err = g.Stabilize(ctx)
	testutil.Error(t, err)
	testutil.Equal(t, 2, called)
	testutil.Equal(t, "", obm.Value())
	testutil.Equal(t, 0, len(cache.cache))

	bv.Set("c")
	err = g.Stabilize(ctx)
	testutil.Error(t, err)
	testutil.Equal(t, 3, called)
	testutil.Equal(t, "", obm.Value())
	testutil.Equal(t, 0, len(cache.cache))
}
