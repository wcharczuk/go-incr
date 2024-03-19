package incrutil

import (
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Accumulate(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, "foo")
	accum := Accumulate(g, v, func(oldValues []string, newValue string) []string {
		return append(oldValues, newValue)
	})
	o := incr.MustObserve(g, accum)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []string{"foo"}, o.Value())

	v.Set("bar")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []string{"foo", "bar"}, o.Value())

	v.Set("baz")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []string{"foo", "bar", "baz"}, o.Value())
}
