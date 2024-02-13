package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_New(t *testing.T) {
	g := New()
	r0 := Return(g, "hello")
	r1 := Return(g, "world!")
	m0 := Map2(g, r0, r1, func(v0, v1 string) string { return v0 + v1 })
	_ = Observe(g, m0)

	testutil.Equal(t, true, g.Has(r0))
	testutil.Equal(t, true, g.Has(r1))
	testutil.Equal(t, true, g.Has(m0))
}

func Test_New_options(t *testing.T) {
	g := New(OptGraphMaxHeight(1024))
	testutil.NotEqual(t, 1024, DefaultMaxHeight)
	testutil.Equal(t, 1024, len(g.recomputeHeap.heights))
	testutil.Equal(t, 1024, len(g.adjustHeightsHeap.nodesByHeight))
}

func Test_Graph_Metadata(t *testing.T) {
	g := New()
	testutil.Nil(t, g.Metadata())
	g.SetMetadata("foo")
	testutil.Equal(t, "foo", g.Metadata())
}

func Test_Graph_Label(t *testing.T) {
	g := New()
	testutil.Equal(t, "", g.Label())
	g.SetLabel("hello")
	testutil.Equal(t, "hello", g.Label())
}

func Test_Graph_IsStabilizing(t *testing.T) {
	g := New()
	testutil.Equal(t, false, g.IsStabilizing())
	g.status = StatusStabilizing
	testutil.Equal(t, true, g.IsStabilizing())
	g.status = StatusNotStabilizing
	testutil.Equal(t, false, g.IsStabilizing())
}

func Test_Graph_addObserver_rediscover(t *testing.T) {
	g := New()

	v := Var(g, "hello")
	o := Observe(g, v)
	_, ok := g.observers[o.Node().ID()]
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 2, g.numNodes)
	testutil.Equal(t, 1, o.Node().height)
	testutil.Equal(t, true, g.recomputeHeap.has(o))
	g.recomputeHeap.remove(o)
	testutil.Equal(t, false, g.recomputeHeap.has(o))

	_ = g.addObserver(o)
	testutil.Equal(t, 2, g.numNodes)
	testutil.Equal(t, 1, o.Node().height)
	testutil.Equal(t, false, g.recomputeHeap.has(o))
}

func Test_Graph_recompute_recomputesObservers(t *testing.T) {
	g := New()
	n := newMockBareNode(g)
	o := Observe(g, n)
	g.recomputeHeap.clear()

	testutil.Equal(t, false, g.recomputeHeap.has(n))
	testutil.Equal(t, false, g.recomputeHeap.has(o))

	err := g.recompute(testContext(), n)
	testutil.Nil(t, err)
	testutil.Equal(t, 1, g.recomputeHeap.len())
	testutil.Equal(t, true, g.recomputeHeap.has(o))
}

func Test_Graph_removeNodeFromGraph(t *testing.T) {
	g := New()

	mn00 := newMockBareNodeWithHeight(2)
	g.numNodes = 2

	g.nodes[mn00.n.id] = mn00

	g.handleAfterStabilization[mn00.n.id] = []func(context.Context){
		func(_ context.Context) {},
		func(_ context.Context) {},
	}
	g.recomputeHeap.add(mn00)
	g.adjustHeightsHeap.add(mn00)

	g.removeNode(mn00)

	testutil.Equal(t, 1, g.numNodes)
	testutil.Equal(t, false, g.recomputeHeap.has(mn00))
	testutil.NoError(t, g.recomputeHeap.sanityCheck())

	testutil.Equal(t, 1, g.numNodes)

	testutil.Equal(t, 0, mn00.n.setAt)
	testutil.Equal(t, 0, mn00.n.boundAt)
	testutil.Equal(t, 0, mn00.n.recomputedAt)
	testutil.Nil(t, mn00.n.createdIn)
	testutil.Nil(t, mn00.n.graph)
	testutil.Equal(t, 0, mn00.n.height)
	testutil.Equal(t, 0, mn00.n.heightInRecomputeHeap)
	testutil.Equal(t, 0, mn00.n.heightInAdjustHeightsHeap)
}
