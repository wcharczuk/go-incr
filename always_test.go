package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Always(t *testing.T) {
	v := Var("foo")
	m0 := Map(v, ident)
	a := Always(m0)
	m1 := Map(a, ident)

	a.(AlwaysIncr[string]).Always() // does nothing

	var updates int
	m1.Node().OnUpdate(func(_ context.Context) {
		updates++
	})

	g := New()
	o := Observe(g, m1)

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
