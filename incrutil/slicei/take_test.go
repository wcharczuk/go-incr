package slicei

import (
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_First(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := First(g, v)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 1, of.Value())
}

func Test_First_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := First(g, v)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 0, of.Value())
}

func Test_Last(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := Last(g, v)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 9, of.Value())
}

func Test_Last_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := Last(g, v)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, 0, of.Value())
}

func Test_TakeFirst(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := TakeFirst(g, v, 5)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0, 1, 2, 3, 4}, of.Value())
}

func Test_TakeFirst_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := TakeFirst(g, v, 5)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{}, of.Value())
}

func Test_TakeLast(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := TakeLast(g, v, 5)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{5, 6, 7, 8, 9}, of.Value())
}

func Test_TakeLast_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := TakeLast(g, v, 5)
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{}, of.Value())
}

func Test_TakeFirstSearch(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := TakeFirstSearch(g, v, func(v int) bool {
		return v >= 5
	})
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0, 1, 2, 3, 4}, of.Value())
}

func Test_TakeFirstSearch_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := TakeFirstSearch(g, v, func(v int) bool {
		return v >= 5
	})
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{}, of.Value())
}

func Test_TakeLastSearch_empty(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{})
	f := TakeLastSearch(g, v, func(v int) bool {
		return v > 5
	})
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{}, of.Value())
}

func Test_TakeLastSearch(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9})
	f := TakeLastSearch(g, v, func(v int) bool {
		return v > 5
	})
	of := incr.MustObserve(g, f)
	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{6, 7, 8, 9}, of.Value())
}
