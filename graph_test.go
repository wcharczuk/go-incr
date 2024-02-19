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
	_ = MustObserve(g, m0)

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

func Test_Graph_Scope(t *testing.T) {
	g := New()

	testutil.NotNil(t, g.scopeGraph())
	testutil.Equal(t, HeightUnset, g.scopeHeight())
	testutil.Equal(t, true, g.isTopScope())
	testutil.Equal(t, true, g.isScopeNecessary())
	testutil.Equal(t, true, g.isScopeValid())

	testutil.Matches(t, `\{graph:(.*)\}`, g.String())
}

func Test_Graph_addObserver_rediscover(t *testing.T) {
	g := New()

	v := Var(g, "hello")
	o := MustObserve(g, v)
	_, ok := g.observers[o.Node().ID()]
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 2, g.numNodes)
	testutil.Equal(t, -1, o.Node().height)
	testutil.Equal(t, false, g.recomputeHeap.has(o))

	g.addObserver(o)
	testutil.Equal(t, 2, g.numNodes)
	testutil.Equal(t, -1, o.Node().height)
	testutil.Equal(t, false, g.recomputeHeap.has(o))
}

func Test_Graph_recompute_recomputesObservers(t *testing.T) {
	g := New()
	n := newMockBareNode(g)
	o := MustObserve(g, n)
	g.recomputeHeap.clear()

	testutil.Equal(t, false, g.recomputeHeap.has(n))
	testutil.Equal(t, false, g.recomputeHeap.has(o))

	err := g.recompute(testContext(), n, false)
	testutil.Nil(t, err)
	testutil.Equal(t, 0, g.recomputeHeap.len())
	testutil.Equal(t, false, g.recomputeHeap.has(o))
}

func Test_Graph_recompute_recomputesObservers_parallel(t *testing.T) {
	g := New()
	n := newMockBareNode(g)
	o := MustObserve(g, n)
	g.recomputeHeap.clear()

	testutil.Equal(t, false, g.recomputeHeap.has(n))
	testutil.Equal(t, false, g.recomputeHeap.has(o))

	err := g.recompute(testContext(), n, true)
	testutil.Nil(t, err)
	testutil.Equal(t, 0, g.recomputeHeap.len())
	testutil.Equal(t, false, g.recomputeHeap.has(o))
}

func Test_Graph_removeNodeFromGraph(t *testing.T) {
	g := New()

	mn00 := newMockBareNodeWithHeight(g, 2)
	g.numNodes = 2

	g.nodes[mn00.n.id] = mn00

	g.handleAfterStabilization[mn00.n.id] = []func(context.Context){
		func(_ context.Context) {},
		func(_ context.Context) {},
	}
	g.recomputeHeap.add(mn00)

	g.removeNode(mn00)

	testutil.Equal(t, 1, g.numNodes)
	testutil.Equal(t, false, g.recomputeHeap.has(mn00))
	testutil.NoError(t, g.recomputeHeap.sanityCheck())

	testutil.Equal(t, 1, g.numNodes)

	testutil.Equal(t, 0, mn00.n.setAt)
	testutil.Equal(t, 0, mn00.n.recomputedAt)
	testutil.NotNil(t, mn00.n.createdIn)
	testutil.Equal(t, HeightUnset, mn00.n.height)
	testutil.Equal(t, HeightUnset, mn00.n.heightInRecomputeHeap)
	testutil.Equal(t, HeightUnset, mn00.n.heightInAdjustHeightsHeap)
}

func Test_Graph_zeroNode(t *testing.T) {
	g := New()

	r := Return(g, "hello")
	_ = MustObserve(g, r)

	testutil.Equal(t, 0, r.Node().height)
	testutil.Equal(t, 0, r.Node().heightInRecomputeHeap)
	testutil.Equal(t, HeightUnset, r.Node().heightInAdjustHeightsHeap)
	testutil.Equal(t, true, r.Node().valid)
	testutil.NotNil(t, r.Node().createdIn)
	testutil.NotEmpty(t, r.Node().observers)

	testutil.Equal(t, 2, g.numNodes)

	r.Node().setAt = 3
	r.Node().changedAt = 4
	r.Node().recomputedAt = 5

	g.zeroNode(r)

	testutil.Equal(t, HeightUnset, r.Node().height)
	testutil.Equal(t, HeightUnset, r.Node().heightInRecomputeHeap)
	testutil.Equal(t, HeightUnset, r.Node().heightInAdjustHeightsHeap)
	testutil.Equal(t, true, r.Node().valid)
	testutil.NotNil(t, r.Node().createdIn)
	testutil.Equal(t, 0, r.Node().setAt)
	testutil.Equal(t, 0, r.Node().changedAt)
	testutil.Equal(t, 0, r.Node().recomputedAt)
	testutil.Empty(t, r.Node().observers)

	testutil.Equal(t, 1, g.numNodes)
}

func Test_Graph_addChild(t *testing.T) {
	g := New()

	n0 := newMockBareNode(g)
	n1 := newMockBareNode(g)

	var err error
	err = g.addChild(nil, n0)
	testutil.Error(t, err)

	err = g.addChild(n0, nil)
	testutil.Error(t, err)

	err = g.addChild(n0, n1)
	testutil.NoError(t, err)
}
