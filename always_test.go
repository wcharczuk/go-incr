package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Always(t *testing.T) {
	v := Var(Root(), "foo")
	m0 := Map(Root(), v, ident)
	a := Always(Root(), m0)
	m1 := Map(Root(), a, ident)

	a.(AlwaysIncr[string]).Always() // does nothing

	var updates int
	m1.Node().OnUpdate(func(_ context.Context) {
		updates++
	})

	g := New()
	o := Observe(Root(), g, m1)

	ctx := testContext()
	_ = g.Stabilize(ctx)

	testutil.ItsEqual(t, "foo", o.Value())
	testutil.ItsEqual(t, 1, updates)

	_ = g.Stabilize(ctx)

	testutil.ItsEqual(t, "foo", o.Value())
	testutil.ItsEqual(t, 2, updates)

	v.Set("bar")

	_ = g.Stabilize(ctx)

	testutil.ItsEqual(t, "bar", o.Value())
	testutil.ItsEqual(t, 3, updates)
}
