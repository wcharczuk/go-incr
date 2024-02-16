package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Observe(t *testing.T) {
	g := New()
	v := Var(g, "foo")
	m0 := Map(g, v, ident)
	o, err := Observe(g, m0)
	testutil.NoError(t, err)

	testutil.Equal(t, 0, v.Node().height)
	testutil.Equal(t, 1, m0.Node().height)
	testutil.Equal(t, -1, o.Node().height)

	ctx := testContext()
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, "foo", o.Value())
}

func Test_Observe_unobserve(t *testing.T) {
	g := New()
	v := Var(g, "foo")
	m0 := Map(g, v, ident)
	o, err := Observe(g, m0)
	testutil.NoError(t, err)

	testutil.Equal(t, 0, v.Node().height)
	testutil.Equal(t, 1, m0.Node().height)
	testutil.Equal(t, -1, o.Node().height)

	ctx := testContext()
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, "foo", o.Value())

	o.Unobserve(ctx)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "", o.Value())

	testutil.Equal(t, -1, v.Node().height)
	testutil.Equal(t, -1, m0.Node().height)
	testutil.Equal(t, -1, o.Node().height)
}

func Test_Observe_unobserve_multi(t *testing.T) {
	g := New()
	v := Var(g, "foo")
	m0 := Map(g, v, ident)
	o0, err := Observe(g, m0)
	testutil.NoError(t, err)
	o1, err := Observe(g, m0)
	testutil.NoError(t, err)

	testutil.Equal(t, 0, v.Node().height)
	testutil.Equal(t, 1, m0.Node().height)
	testutil.Equal(t, -1, o0.Node().height)
	testutil.Equal(t, -1, o1.Node().height)

	ctx := testContext()
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, "foo", o0.Value())
	testutil.Equal(t, "foo", o1.Value())

	o0.Unobserve(ctx)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "", o0.Value())
	testutil.Equal(t, "foo", o1.Value())

	testutil.Equal(t, 0, v.Node().height)
	testutil.Equal(t, 1, m0.Node().height)
	testutil.Equal(t, -1, o0.Node().height)
	testutil.Equal(t, -1, o1.Node().height)
}

func Test_Observe_unobserve_var(t *testing.T) {
	g := New()
	v := Var(g, "foo")
	m0 := Map(g, v, ident)
	o, err := Observe(g, m0)
	testutil.NoError(t, err)

	testutil.Equal(t, 0, v.Node().height)
	testutil.Equal(t, 1, m0.Node().height)
	testutil.Equal(t, -1, o.Node().height)

	ctx := testContext()
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, "foo", o.Value())

	o.Unobserve(ctx)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "", o.Value())

	testutil.NotNil(t, v.Node().createdIn)
	testutil.Equal(t, -1, v.Node().height)
	testutil.Equal(t, -1, m0.Node().height)
	testutil.Equal(t, -1, o.Node().height)

	o2, err := Observe(g, m0)
	testutil.NoError(t, err)

	testutil.Equal(t, 0, v.Node().height)
	testutil.Equal(t, 1, m0.Node().height)
	testutil.Equal(t, -1, o.Node().height)
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "foo", o2.Value())
}
