package slicei

import (
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_First(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := First(g, v, 5)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0, 1, 2, 3, 4}, of.Value())
}

func Test_First_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := First(g, v, 5)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{}, of.Value())
}

func Test_Last(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := Last(g, v, 5)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{5, 6, 7, 8, 9}, of.Value())
}

func Test_Last_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := Last(g, v, 5)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{}, of.Value())
}

func Test_BeforeSorted(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := BeforeSorted(g, v, func(v int) bool {
		return v >= 5
	})
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0, 1, 2, 3, 4}, of.Value())
}

func Test_BeforeSorted_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := BeforeSorted(g, v, func(v int) bool {
		return v >= 5
	})
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{}, of.Value())
}

func Test_AfterSorted_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := AfterSorted(g, v, func(v int) bool {
		return v > 5
	})
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{}, of.Value())
}

func Test_AfterSorted(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := AfterSorted(g, v, func(v int) bool {
		return v > 5
	})
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{6, 7, 8, 9}, of.Value())
}
