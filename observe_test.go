package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Observe(t *testing.T) {
	g := New()
	v := Var(g, "foo")
	m0 := Map(g, v, ident)
	o := Observe(g, m0)

	testutil.Equal(t, 0, v.Node().height)
	testutil.Equal(t, 1, m0.Node().height)
	testutil.Equal(t, 2, o.Node().height)

	ctx := testContext()
	_ = g.Stabilize(ctx)

	testutil.Equal(t, "foo", o.Value())
}
