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
