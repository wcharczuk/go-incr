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
