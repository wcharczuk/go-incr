package slicei

import (
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Filter(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	f := Filter(g, v, func(val int) bool {
		return val%2 == 0
	})
	of := incr.MustObserve(g, f)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0, 2, 4, 6, 8, 10}, of.Value())
}
