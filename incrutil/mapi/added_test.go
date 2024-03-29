package mapi

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Added(t *testing.T) {
	ctx := context.Background()
	g := incr.New()
	v := incr.Var(g, map[string]any{"foo": 1, "bar": 2})

	d := Added(g, v)

	od := incr.MustObserve(g, d)

	_ = g.Stabilize(ctx)

	testutil.Equal(t, 2, len(od.Value()))

	v.Set(map[string]any{"foo": 1, "bar": 2, "snoo": 3})
	_ = g.Stabilize(ctx)

	testutil.Equal(t, 1, len(od.Value()))
	testutil.Equal(t, 3, od.Value()["snoo"])

	v.Set(map[string]any{"foo": 2, "bar": 3, "snoo": 4})
	_ = g.Stabilize(ctx)

	testutil.Equal(t, 0, len(od.Value()))
}
