package incrutil

import (
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
	bm.Purge("a")

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
	bm.Clear()

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
