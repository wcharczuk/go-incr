package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_NodeStats(t *testing.T) {
	v0 := Var("a")
	v0.Node().SetLabel("v0")

	av := Var("a-value")
	a0 := Map(av, ident)
	a1 := Map(a0, ident)

	bv := Var("b-value")
	b0 := Map(bv, ident)
	b1 := Map(b0, ident)
	b2 := Map(b1, ident)

	bind := Bind(v0, func(which string) Incr[string] {
		if which == "a" {
			return a1
		}
		if which == "b" {
			return b2
		}
		return nil
	})

	s0 := Return("hello")
	s1 := Map(s0, ident)
	o := Map2(bind, s1, concat)

	g := New()
	_ = Observe(g, o)

	_ = g.Stabilize(context.TODO())

	vStats := NodeStats(v0)
	testutil.ItsEqual(t, 1, vStats.Changes())
	testutil.ItsEqual(t, 1, vStats.Recomputes())
	testutil.ItsEqual(t, 1, vStats.Children())
	testutil.ItsEqual(t, 0, vStats.Parents())
	testutil.ItsEqual(t, 0, vStats.SetAt())
	testutil.ItsEqual(t, 1, vStats.ChangedAt())

	oStats := NodeStats(o)
	testutil.ItsEqual(t, 1, oStats.Changes())
	testutil.ItsEqual(t, 1, oStats.Recomputes())
	testutil.ItsEqual(t, 1, oStats.Children())
	testutil.ItsEqual(t, 3, oStats.Parents())
	testutil.ItsEqual(t, 0, oStats.SetAt())
	testutil.ItsEqual(t, 1, oStats.ChangedAt())

	v0.Set("b")
	_ = g.Stabilize(context.TODO())

	vStats = NodeStats(v0)
	testutil.ItsEqual(t, 2, vStats.Changes())
	testutil.ItsEqual(t, 2, vStats.Recomputes())
	testutil.ItsEqual(t, 1, vStats.Children())
	testutil.ItsEqual(t, 0, vStats.Parents())
	testutil.ItsEqual(t, 2, vStats.SetAt())
	testutil.ItsEqual(t, 2, vStats.ChangedAt())

	oStats = NodeStats(o)
	testutil.ItsEqual(t, 2, oStats.Changes())
	testutil.ItsEqual(t, 2, oStats.Recomputes())
	testutil.ItsEqual(t, 1, oStats.Children())
	testutil.ItsEqual(t, 3, oStats.Parents())
	testutil.ItsEqual(t, 0, oStats.SetAt())
	testutil.ItsEqual(t, 2, oStats.ChangedAt())
}

func Test_GraphStats(t *testing.T) {
	gs := graphStats{
		stabilizationNum:   4,
		numNodes:           3,
		numNodesRecomputed: 2,
		numNodesChanged:    1,
	}

	testutil.ItsEqual(t, 4, gs.StabilizationNum())
	testutil.ItsEqual(t, 3, gs.Nodes())
	testutil.ItsEqual(t, 2, gs.NodesRecomputed())
	testutil.ItsEqual(t, 1, gs.NodesChanged())
}

func Test_Graph_Stats(t *testing.T) {
	g := &Graph{
		numNodes:           3,
		numNodesRecomputed: 2,
		numNodesChanged:    1,
		recomputeHeap:      new(recomputeHeap),
	}
	gs := g.Stats()

	testutil.ItsEqual(t, 3, gs.Nodes())
	testutil.ItsEqual(t, 2, gs.NodesRecomputed())
	testutil.ItsEqual(t, 1, gs.NodesChanged())
	testutil.ItsEqual(t, 0, gs.RecomputeHeapLength())
}
