package incrutil

import (
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_CutoffUnchanged(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, "hello")
	cov := CutoffUnchanged(g, v)
	ocov := incr.MustObserve(g, cov)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "hello", ocov.Value())

	testutil.Equal(t, 1, incr.ExpertNode(cov).NumChanges())

	v.Set("not-hello")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "not-hello", ocov.Value())

	testutil.Equal(t, 2, incr.ExpertNode(cov).NumRecomputes())
	testutil.Equal(t, 2, incr.ExpertNode(cov).NumChanges())

	v.Set("not-hello")

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, "not-hello", ocov.Value())

	testutil.Equal(t, 3, incr.ExpertNode(cov).NumRecomputes())
	testutil.Equal(t, 2, incr.ExpertNode(cov).NumChanges())
}
