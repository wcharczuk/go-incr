package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Always(t *testing.T) {
	g := New()
	v := Var(g, "foo")
	v.Node().SetLabel("v")
	m0 := Map(g, v, ident)
	m0.Node().SetLabel("m0")
	a := Always(g, m0)
	a.Node().SetLabel("a")
	m1 := Map(g, a, ident)
	m1.Node().SetLabel("m1")

	a.(AlwaysIncr[string]).Always() // does nothing

	var updates int
	m1.Node().OnUpdate(func(_ context.Context) {
		updates++
	})
	o := MustObserve(g, m1)

	testutil.Equal(t, 0, v.Node().height)
	testutil.Equal(t, 1, m0.Node().height)
	testutil.Equal(t, 2, a.Node().height)
	testutil.Equal(t, 3, m1.Node().height)
	testutil.Equal(t, -1, o.Node().height)

	ctx := testContext()
	_ = g.Stabilize(ctx)

	testutil.Equal(t, "foo", o.Value())
	testutil.Equal(t, 1, updates)

	_ = g.Stabilize(ctx)

	testutil.Equal(t, "foo", o.Value())
	testutil.Equal(t, 2, updates)

	v.Set("bar")

	_ = g.Stabilize(ctx)

	testutil.Equal(t, "bar", o.Value())
	testutil.Equal(t, 3, updates)
}
