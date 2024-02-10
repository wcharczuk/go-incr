package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Var_Set_unobserved(t *testing.T) {
	v := Var(Root(), "foo")

	testutil.Equal(t, "foo", v.Value())

	v.Set("not-foo")

	testutil.Equal(t, "not-foo", v.Value())
}

func Test_Var_Stabilize_zero(t *testing.T) {
	v := Var(Root(), "foo")

	g := New()
	_ = Observe(Root(), g, v)

	_ = g.Stabilize(context.TODO())
	testutil.Equal(t, "foo", v.Value())
}

func Test_Var_Set_duringStabilization(t *testing.T) {
	v := Var(Root(), "foo")
	g := New()
	_ = Observe(Root(), g, v)
	g.status = StatusStabilizing

	v.Set("not-foo")

	testutil.Equal(t, true, v.(*varIncr[string]).setDuringStabilization)
	testutil.Equal(t, "not-foo", v.(*varIncr[string]).setDuringStabilizationValue)
	testutil.Equal(t, "foo", v.(*varIncr[string]).value)

	_ = v.(*varIncr[string]).Stabilize(context.TODO())

	testutil.Equal(t, false, v.(*varIncr[string]).setDuringStabilization)
	testutil.Equal(t, "", v.(*varIncr[string]).setDuringStabilizationValue)
	testutil.Equal(t, "not-foo", v.(*varIncr[string]).value)
}
