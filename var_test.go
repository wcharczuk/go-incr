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

func Test_Var_Set_duringStabilization(t *testing.T) {
	v := Var("foo")
	g := New()
	_ = Observe(g, v)
	g.status = StatusStabilizing

	v.Set("not-foo")

	testutil.ItsEqual(t, true, v.(*varIncr[string]).setDuringStabilization)
	testutil.ItsEqual(t, "not-foo", v.(*varIncr[string]).setDuringStabilizationValue)
	testutil.ItsEqual(t, "foo", v.(*varIncr[string]).value)

	_ = v.(*varIncr[string]).Stabilize(context.TODO())

	testutil.ItsEqual(t, false, v.(*varIncr[string]).setDuringStabilization)
	testutil.ItsEqual(t, "", v.(*varIncr[string]).setDuringStabilizationValue)
	testutil.ItsEqual(t, "not-foo", v.(*varIncr[string]).value)
}
