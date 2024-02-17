package incr

import (
	"fmt"
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

	testutil.Matches(t, `observer\[(.*)\]`, fmt.Sprint(o))
	o.Node().SetLabel("foo")
	testutil.Matches(t, `observer\[(.*)\]:foo`, fmt.Sprint(o))
}

func Test_Observe_error(t *testing.T) {
	g := New(OptGraphMaxHeight(4))
	v := Var(g, "foo")
	m0 := Map(g, v, ident)
	m1 := Map(g, m0, ident)
	m2 := Map(g, m1, ident)
	m3 := Map(g, m2, ident)
	o, err := Observe(g, m3)
	testutil.Nil(t, o)
	testutil.Error(t, err)
}

func Test_MustObserve_panic(t *testing.T) {
	g := New(OptGraphMaxHeight(4))
	v := Var(g, "foo")
	m0 := Map(g, v, ident)
	m1 := Map(g, m0, ident)
	m2 := Map(g, m1, ident)
	m3 := Map(g, m2, ident)

	var recovered any
	func() {
		defer func() {
			recovered = recover()
		}()
		_ = MustObserve(g, m3)

	}()
	testutil.NotNil(t, recovered)
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

func Test_Observe_alreadyNecessary(t *testing.T) {
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

	o2 := MustObserve(g, m0)
	err = g.Stabilize(ctx)
	testutil.NoError(t, err)

	testutil.Equal(t, "foo", o2.Value())
}
