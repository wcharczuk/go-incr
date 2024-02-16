package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Var_Set_unobserved(t *testing.T) {
	g := New()
	v := Var(g, "foo")

	testutil.Equal(t, "foo", v.Value())

	v.Set("not-foo")

	testutil.Equal(t, "not-foo", v.Value())
}

func Test_Var_Stabilize_zero(t *testing.T) {
	g := New()
	v := Var(g, "foo")

	_ = MustObserve(g, v)

	_ = g.Stabilize(context.TODO())
	testutil.Equal(t, "foo", v.Value())
}

func Test_Var_Set_duringStabilization(t *testing.T) {
	g := New()
	v := Var(g, "foo")
	_ = MustObserve(g, v)
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

func Test_Var_Set_duringStabilization_realistic(t *testing.T) {
	ctx := testContext()
	g := New()
	v := Var(g, "foo")
	proceed := make(chan struct{})
	invoked := make(chan struct{})
	m0 := Map(g, v, func(vv string) string {
		close(invoked)
		<-proceed
		return vv + "-done!"
	})
	o := MustObserve(g, m0)

	stabilizationDone := make(chan struct{})
	go func() {
		_ = g.Stabilize(ctx)
		close(stabilizationDone)
	}()
	<-invoked
	v.Set("during-stab")
	close(proceed)
	<-stabilizationDone
	testutil.Equal(t, "foo-done!", o.Value())

	proceed = make(chan struct{})
	invoked = make(chan struct{})
	stabilizationDone = make(chan struct{})
	go func() {
		_ = g.Stabilize(ctx)
		close(stabilizationDone)
	}()
	<-invoked
	close(proceed)
	<-stabilizationDone
	testutil.Equal(t, "during-stab-done!", o.Value())
}

func Test_Var_ShouldBeInvalidated(t *testing.T) {
	g := New()
	v := Var(g, "foo")

	testutil.Equal(t, false, v.ShouldBeInvalidated())
}
