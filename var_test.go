package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Var_Set_unobserved(t *testing.T) {
	v := Var(testContext(), "foo")

	testutil.ItsEqual(t, "foo", v.Value())

	v.Set("not-foo")

	testutil.ItsEqual(t, "not-foo", v.Value())
}

func Test_Var_Stabilize_zero(t *testing.T) {
	ctx := testContext()
	v := Var(ctx, "foo")

	g := New()
	_ = Observe(ctx, g, v)

	_ = g.Stabilize(context.TODO())
	testutil.ItsEqual(t, "foo", v.Value())
}

func Test_Var_Set_duringStabilization(t *testing.T) {
	ctx := testContext()
	v := Var(testContext(), "foo")
	g := New()
	_ = Observe(ctx, g, v)
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
