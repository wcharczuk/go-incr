package mapi

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Removed(t *testing.T) {
	ctx := context.Background()
	g := incr.New()
	v := incr.Var(g, map[string]any{"foo": 1, "bar": 2, "snoo": 3})

	d := Removed(g, v)

	od := incr.MustObserve(g, d)

	_ = g.Stabilize(ctx)

	testutil.Equal(t, 0, len(od.Value()))

	v.Set(map[string]any{"foo": 1, "bar": 2})
	_ = g.Stabilize(ctx)

	testutil.Equal(t, 1, len(od.Value()))
	testutil.Equal(t, 3, od.Value()["snoo"])

	v.Set(map[string]any{"foo": 1, "bar": 2})
	_ = g.Stabilize(ctx)

	testutil.Equal(t, 0, len(od.Value()))
}
