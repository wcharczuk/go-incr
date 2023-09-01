package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Var_Set_unobserved(t *testing.T) {
	v := Var("foo")

	testutil.ItsEqual(t, "foo", v.(*varIncr[string]).setValue)
	testutil.ItsEqual(t, "", v.(*varIncr[string]).value)

	v.Set("not-foo")

	testutil.ItsEqual(t, "foo", v.(*varIncr[string]).setValue)
	testutil.ItsEqual(t, "not-foo", v.(*varIncr[string]).value)
}

func Test_Var_Stabilize_zero(t *testing.T) {
	v := Var("foo")

	g := New()
	_ = Observe(g, v)

	_ = g.Stabilize(context.TODO())
	testutil.ItsEqual(t, "", v.(*varIncr[string]).setValue)
}
