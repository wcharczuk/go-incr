package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Always(t *testing.T) {
	g := New()
	v := Var(g, "foo")
	m0 := Map(g, v, ident)
	a := Always(g, m0)
	m1 := Map(g, a, ident)

	a.(AlwaysIncr[string]).Always() // does nothing

	var updates int
	m1.Node().OnUpdate(func(_ context.Context) {
		updates++
	})
	o := Observe(g, m1)

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
