package incrutil

import (
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_MapLast(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, 1)
	c := MapLast(g, v, func(a0, a1 int) bool { return a0 < a1 })
	oc := incr.MustObserve(g, c)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, true, oc.Value())

	v.Set(2)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, true, oc.Value())

	v.Set(4)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, true, oc.Value())

	v.Set(3)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, false, oc.Value())

	v.Set(2)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, false, oc.Value())

	v.Set(5)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, true, oc.Value())
}
