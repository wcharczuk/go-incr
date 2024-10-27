package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_MapN_AddInput_beforeObservationStabilization(t *testing.T) {
	ctx := testContext()
	g := New()

	r0 := Return(g, 1)
	r1 := Return(g, 2)
	mn := MapN(g, sum, r0, r1)

	r2 := Return(g, 3)
	err := mn.AddInput(r2)
	testutil.NoError(t, err)

	om := MustObserve(g, mn)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 6, om.Value())
}

func Test_MapN_AddInput_afterObservatino(t *testing.T) {
	ctx := testContext()
	g := New()

	r0 := Return(g, 1)
	r1 := Return(g, 2)
	mn := MapN(g, sum, r0, r1)
	om := MustObserve(g, mn)

	r2 := Return(g, 3)
	err := mn.AddInput(r2)
	testutil.NoError(t, err)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 6, om.Value())
}

func Test_MapN_AddInput_afterStabilization(t *testing.T) {
	ctx := testContext()
	g := New()

	r0 := Return(g, 1)
	r1 := Return(g, 2)
	r2 := Return(g, 3)
	mn := MapN(g, sum, r0, r1, r2)
	om := MustObserve(g, mn)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 6, om.Value())

	r3 := Return(g, 4)
	err = mn.AddInput(r3)
	testutil.NoError(t, err)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 10, om.Value())
}

func sum[A ~int | ~float64](values ...A) (out A) {
	for _, v := range values {
		out += v
	}
	return
}

func Test_MapN_RemoveInput(t *testing.T) {
	ctx := testContext()
	g := New()

	r0 := Return(g, 1)
	r1 := Return(g, 2)
	mn := MapN(g, sum, r0, r1)
	om := MustObserve(g, mn)

	r2 := Return(g, 3)
	err := mn.AddInput(r2)
	testutil.NoError(t, err)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 6, om.Value())

	err = mn.RemoveInput(r1.Node().ID())
	testutil.NoError(t, err)

	testutil.Equal(t, 2, len(mn.Node().parents))

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 4, om.Value())

	hasR1 := g.Has(r1)
	testutil.Equal(t, false, hasR1)
}

func Test_MapN_RemoveInput_onlyInput(t *testing.T) {
	ctx := testContext()
	g := New()

	mn := MapN[int](g, sum)
	om := MustObserve(g, mn)

	r2 := Return(g, 3)
	err := mn.AddInput(r2)
	testutil.NoError(t, err)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 3, om.Value())

	err = mn.RemoveInput(r2.Node().ID())
	testutil.NoError(t, err)

	testutil.Equal(t, 0, len(mn.Node().parents))

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 0, om.Value())

	hasR2 := g.Has(r2)
	testutil.Equal(t, false, hasR2)
}
