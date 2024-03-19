package slicei

import (
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Sort_asc(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{4, 3, 5, 2, 1})
	s := Sort(g, v, Asc)
	os := incr.MustObserve(g, s)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{1, 2, 3, 4, 5}, os.Value())
}

func Test_Sort_desc(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{4, 3, 5, 2, 1})
	s := Sort(g, v, Desc)
	os := incr.MustObserve(g, s)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{5, 4, 3, 2, 1}, os.Value())
}
