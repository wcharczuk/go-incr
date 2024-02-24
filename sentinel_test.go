package incr

import (
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Sentinel(t *testing.T) {
	ctx := testContext()
	g := New()
	v := Var(g, "foo")
	var updates int
	m := Map(g, v, func(vv string) string {
		updates++
		return vv + fmt.Sprintf("-mapped-%d", updates)
	})
	s := Sentinel(g, func() bool {
		return updates < 3
	}, m)

	testutil.Equal(t, false, m.Node().isNecessary(), "sentinels should not mark nodes as necessary")

	testutil.NotNil(t, s)

	om := MustObserve(g, m)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-1", om.Value())

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-2", om.Value())

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-3", om.Value())

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-3", om.Value())

	testutil.Equal(t, 3, updates)
}

func Test_Sentinel_Unwatch(t *testing.T) {
	ctx := testContext()
	g := New()
	v := Var(g, "foo")
	var updates int
	m := Map(g, v, func(vv string) string {
		updates++
		return vv + fmt.Sprintf("-mapped-%d", updates)
	})
	s := Sentinel(g, func() bool {
		return updates < 3
	}, m)

	testutil.NotNil(t, s)

	om := MustObserve(g, m)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-1", om.Value())

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-2", om.Value())

	testutil.Equal(t, 2, updates)

	s.Unwatch(ctx)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-2", om.Value())

	testutil.Equal(t, 2, updates)
}

func Test_Sentinel_withinBind(t *testing.T) {
	ctx := testContext()
	g := New()

	v := Var(g, "foo")
	var updates int
	m := Map(g, v, func(vv string) string {
		updates++
		return vv + fmt.Sprintf("-mapped-%d", updates)
	})
	_ = Sentinel(g, func() bool {
		return updates < 3
	}, m)

	bv := Var(g, "a")
	b := Bind(g, bv, func(bs Scope, which string) Incr[string] {
		if which == "a" {
			return m
		}
		return Return(bs, "nope")
	})
	ob := MustObserve(g, b)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-1", ob.Value())

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-2", ob.Value())

	bv.Set("b")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "nope", ob.Value())

	bv.Set("a")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-4", ob.Value())

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo-mapped-4", ob.Value())

	testutil.Equal(t, 4, updates)
}
