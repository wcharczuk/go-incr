package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Var_Set_unobserved(t *testing.T) {
	v := Var("foo")

	testutil.ItsEqual(t, "foo", v.Value())

	v.Set("not-foo")

	testutil.ItsEqual(t, "not-foo", v.Value())
}

func Test_Var_Stabilize_zero(t *testing.T) {
	v := Var("foo")

	g := New()
	_ = Observe(g, v)

	_ = g.Stabilize(context.TODO())
	testutil.ItsEqual(t, "foo", v.Value())
}
