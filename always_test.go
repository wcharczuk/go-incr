package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Always(t *testing.T) {
	ctx := testContext()
	v := Var(ctx, "foo")
	m0 := Map(ctx, v, ident)
	a := Always(ctx, m0)
	m1 := Map(ctx, a, ident)

	a.(AlwaysIncr[string]).Always() // does nothing

	var updates int
	m1.Node().OnUpdate(func(_ context.Context) {
		updates++
	})

	g := New()
	o := Observe(ctx, g, m1)

	_ = g.Stabilize(context.TODO())

	testutil.ItsEqual(t, "foo", o.Value())
	testutil.ItsEqual(t, 1, updates)

	_ = g.Stabilize(context.TODO())

	testutil.ItsEqual(t, "foo", o.Value())
	testutil.ItsEqual(t, 2, updates)

	v.Set("bar")

	_ = g.Stabilize(context.TODO())

	testutil.ItsEqual(t, "bar", o.Value())
	testutil.ItsEqual(t, 3, updates)
}
