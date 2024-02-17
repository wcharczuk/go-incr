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
