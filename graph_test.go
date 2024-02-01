package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_New(t *testing.T) {
	ctx := testContext()
	r0 := Return(ctx, "hello")
	r1 := Return(ctx, "world!")
	m0 := Map2(ctx, r0, r1, func(v0, v1 string) string { return v0 + v1 })
	g := New()
	_ = Observe(g, m0)

	testutil.ItsEqual(t, true, g.IsObserving(r0))
	testutil.ItsEqual(t, true, g.IsObserving(r1))
	testutil.ItsEqual(t, true, g.IsObserving(m0))

	m1 := Map2(ctx, r0, r1, func(v0, v1 string) string { return v0 + v1 })
	testutil.ItsEqual(t, false, g.IsObserving(m1))
}

func Test_New_options(t *testing.T) {
	g := New(GraphMaxRecomputeHeapHeight(1024))
	testutil.ItsNotEqual(t, 1024, DefaultMaxRecomputeHeapHeight)
	testutil.ItsEqual(t, 1024, len(g.recomputeHeap.heights))
}

func Test_Graph_Metadata(t *testing.T) {
	g := New()
	testutil.ItsNil(t, g.Metadata())
	g.SetMetadata("foo")
	testutil.ItsEqual(t, "foo", g.Metadata())
}

func Test_Graph_Label(t *testing.T) {
	g := New()
	testutil.ItsEqual(t, "", g.Label())
	g.SetLabel("hello")
	testutil.ItsEqual(t, "hello", g.Label())
}

func Test_Graph_UnobserveNodes(t *testing.T) {
	ctx := testContext()

	r0 := Return(ctx, "hello")
	m0 := Map(ctx, r0, ident)
	m1 := Map(ctx, m0, ident)
	m2 := Map(ctx, m1, ident)

	ar0 := Return(ctx, "hello")
	am0 := Map(ctx, ar0, ident)
	am1 := Map(ctx, am0, ident)
	am2 := Map(ctx, am1, ident)

	g := New()
	o1 := Observe(g, m1)
	_ = Observe(g, am2)

	testutil.ItsEqual(t, true, g.IsObserving(r0))
	testutil.ItsEqual(t, true, g.IsObserving(m0))
	testutil.ItsEqual(t, true, g.IsObserving(m1))
	testutil.ItsEqual(t, false, g.IsObserving(m2), "using the Observe incremental we actually don't care about m2!")

	testutil.ItsEqual(t, true, g.IsObserving(ar0))
	testutil.ItsEqual(t, true, g.IsObserving(am0))
	testutil.ItsEqual(t, true, g.IsObserving(am1))
	testutil.ItsEqual(t, true, g.IsObserving(am2))

	g.unobserveNodes(ctx, m1, o1)

	testutil.ItsEqual(t, false, g.IsObserving(r0))
	testutil.ItsEqual(t, false, g.IsObserving(m0))
	testutil.ItsEqual(t, false, g.IsObserving(m1))
	testutil.ItsEqual(t, false, g.IsObserving(m2))

	testutil.ItsNil(t, r0.Node().graph)
	testutil.ItsNil(t, m0.Node().graph)
	testutil.ItsNil(t, m1.Node().graph)
	testutil.ItsNil(t, m2.Node().graph)

	testutil.ItsEqual(t, true, g.IsObserving(ar0))
	testutil.ItsEqual(t, true, g.IsObserving(am0))
	testutil.ItsEqual(t, true, g.IsObserving(am1))
	testutil.ItsEqual(t, true, g.IsObserving(am2))
}

func Test_Graph_UnobserveNodes_notObserving(t *testing.T) {
	ctx := testContext()

	r0 := Return(ctx, "hello")
	m0 := Map(ctx, r0, ident)
	m1 := Map(ctx, m0, ident)
	m2 := Map(ctx, m1, ident)

	ar0 := Return(ctx, "hello")
	am0 := Map(ctx, ar0, ident)
	am1 := Map(ctx, am0, ident)
	am2 := Map(ctx, am1, ident)

	g := New()
	o := Observe(g, m1)

	testutil.ItsEqual(t, true, g.IsObserving(r0))
	testutil.ItsEqual(t, true, g.IsObserving(m0))
	testutil.ItsEqual(t, true, g.IsObserving(m1))
	testutil.ItsEqual(t, false, g.IsObserving(m2), "we observed m1, which is the parent of m2!")

	testutil.ItsEqual(t, false, g.IsObserving(ar0))
	testutil.ItsEqual(t, false, g.IsObserving(am0))
	testutil.ItsEqual(t, false, g.IsObserving(am1))
	testutil.ItsEqual(t, false, g.IsObserving(am2))

	g.unobserveNodes(ctx, am1, o)

	testutil.ItsEqual(t, true, g.IsObserving(r0))
	testutil.ItsEqual(t, true, g.IsObserving(m0))
	testutil.ItsEqual(t, true, g.IsObserving(m1))
	testutil.ItsEqual(t, false, g.IsObserving(m2))
}

func Test_Graph_IsStabilizing(t *testing.T) {
	g := New()
	testutil.ItsEqual(t, false, g.IsStabilizing())
	g.status = StatusStabilizing
	testutil.ItsEqual(t, true, g.IsStabilizing())
	g.status = StatusNotStabilizing
	testutil.ItsEqual(t, false, g.IsStabilizing())
}

func Test_Graph_recomputeHeights(t *testing.T) {
	g := New()

	n0 := emptyNode{NewNode()}
	n1 := emptyNode{NewNode()}
	n2 := emptyNode{NewNode()}
	n3 := emptyNode{NewNode()}

	Link(n1, n0)
	Link(n2, n1)
	Link(n3, n2)

	g.recomputeHeights(n1)

	testutil.ItsEqual(t, 0, n0.n.height)
	testutil.ItsEqual(t, 2, n1.n.height)
	testutil.ItsEqual(t, 3, n2.n.height)
	testutil.ItsEqual(t, 4, n3.n.height)

	testutil.ItsEqual(t, 1, n0.n.numComputePseudoheights)
	testutil.ItsEqual(t, 1, n1.n.numComputePseudoheights)
	testutil.ItsEqual(t, 1, n2.n.numComputePseudoheights)
	testutil.ItsEqual(t, 1, n3.n.numComputePseudoheights)
}

func Test_Graph_recomputeHeights_observed(t *testing.T) {
	ctx := testContext()
	g := New()

	v0 := Var(ctx, "a")
	m0 := Map(ctx, v0, ident)
	o0 := Observe(g, m0)

	m1 := Map(ctx, m0, ident)
	m2 := Map(ctx, m1, ident)
	o1 := Observe(g, m2)

	m0.Node().height = 1
	g.recomputeHeights(m0)

	err := g.Stabilize(ctx)
	testutil.ItsNil(t, err)

	testutil.ItsEqual(t, "a", o0.Value())
	testutil.ItsEqual(t, "a", o1.Value())
}

func Test_Graph_addObserver_rediscover(t *testing.T) {
	ctx := testContext()
	g := New()

	v := Var(ctx, "hello")
	o := Observe(g, v)
	_, ok := g.observers[o.Node().ID()]
	testutil.ItsEqual(t, true, ok)
	testutil.ItsEqual(t, 2, g.numNodes)
	testutil.ItsEqual(t, 2, o.Node().height)
	testutil.ItsEqual(t, true, g.recomputeHeap.Has(o))
	g.recomputeHeap.Remove(o)
	testutil.ItsEqual(t, false, g.recomputeHeap.Has(o))

	g.addObserver(ctx, o)
	testutil.ItsEqual(t, 2, g.numNodes)
	testutil.ItsEqual(t, 2, o.Node().height)
	testutil.ItsEqual(t, false, g.recomputeHeap.Has(o))
}

func Test_Graph_recompute_nilNodeMetadata(t *testing.T) {
	g := New()

	n := newMockBareNode()
	n.n = nil
	err := g.recompute(testContext(), n)
	testutil.ItsNotNil(t, err)
}

func Test_Node_computePseudoHeight(t *testing.T) {
	c010 := newMockBareNode()
	c10 := newMockBareNode()
	c00 := newMockBareNode()
	c01 := newMockBareNode()
	c0 := newMockBareNode()
	c1 := newMockBareNode()
	c2 := newMockBareNode()
	p := newMockBareNode()

	Link(c01, c010)
	Link(c0, c00, c01)
	Link(c1, c10)
	Link(p, c0, c1, c2)

	graph := New()

	testutil.ItsEqual(t, 4, graph.computePseudoHeight(make(map[Identifier]int), p))
	testutil.ItsEqual(t, 3, graph.computePseudoHeight(make(map[Identifier]int), c0))
	testutil.ItsEqual(t, 2, graph.computePseudoHeight(make(map[Identifier]int), c1))
	testutil.ItsEqual(t, 1, graph.computePseudoHeight(make(map[Identifier]int), c2))
}
