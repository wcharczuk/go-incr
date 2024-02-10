package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_New(t *testing.T) {
	r0 := Return(Root(), "hello")
	r1 := Return(Root(), "world!")
	m0 := Map2(Root(), r0, r1, func(v0, v1 string) string { return v0 + v1 })
	g := New()
	_ = Observe(Root(), g, m0)

	testutil.Equal(t, true, g.IsObserving(r0))
	testutil.Equal(t, true, g.IsObserving(r1))
	testutil.Equal(t, true, g.IsObserving(m0))

	m1 := Map2(Root(), r0, r1, func(v0, v1 string) string { return v0 + v1 })
	testutil.Equal(t, false, g.IsObserving(m1))
}

func Test_New_options(t *testing.T) {
	g := New(GraphMaxRecomputeHeapHeight(1024))
	testutil.NotEqual(t, 1024, DefaultMaxHeight)
	testutil.Equal(t, 1024, len(g.recomputeHeap.heights))
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

func Test_Graph_UnobserveNodes(t *testing.T) {
	ctx := testContext()

	r0 := Return(Root(), "hello")
	m0 := Map(Root(), r0, ident)
	m1 := Map(Root(), m0, ident)
	m2 := Map(Root(), m1, ident)

	ar0 := Return(Root(), "hello")
	am0 := Map(Root(), ar0, ident)
	am1 := Map(Root(), am0, ident)
	am2 := Map(Root(), am1, ident)

	g := New()
	o1 := Observe(Root(), g, m1)
	_ = Observe(Root(), g, am2)

	testutil.Equal(t, true, g.IsObserving(r0))
	testutil.Equal(t, true, g.IsObserving(m0))
	testutil.Equal(t, true, g.IsObserving(m1))
	testutil.Equal(t, false, g.IsObserving(m2), "using the Observe incremental we actually don't care about m2!")

	testutil.Equal(t, true, g.IsObserving(ar0))
	testutil.Equal(t, true, g.IsObserving(am0))
	testutil.Equal(t, true, g.IsObserving(am1))
	testutil.Equal(t, true, g.IsObserving(am2))

	Unlink(o1, m1)
	g.unobserveNodes(ctx, m1, o1)

	testutil.Equal(t, false, g.IsObserving(r0))
	testutil.Equal(t, false, g.IsObserving(m0))
	testutil.Equal(t, false, g.IsObserving(m1))
	testutil.Equal(t, false, g.IsObserving(m2))

	testutil.Nil(t, r0.Node().graph)
	testutil.Nil(t, m0.Node().graph)
	testutil.Nil(t, m1.Node().graph)
	testutil.Nil(t, m2.Node().graph)

	testutil.Equal(t, true, g.IsObserving(ar0))
	testutil.Equal(t, true, g.IsObserving(am0))
	testutil.Equal(t, true, g.IsObserving(am1))
	testutil.Equal(t, true, g.IsObserving(am2))
}

func Test_Graph_UnobserveNodes_notObserving(t *testing.T) {
	ctx := testContext()

	r0 := Return(Root(), "hello")
	m0 := Map(Root(), r0, ident)
	m1 := Map(Root(), m0, ident)
	m2 := Map(Root(), m1, ident)

	ar0 := Return(Root(), "hello")
	am0 := Map(Root(), ar0, ident)
	am1 := Map(Root(), am0, ident)
	am2 := Map(Root(), am1, ident)

	g := New()
	o := Observe(Root(), g, m1)

	testutil.Equal(t, true, g.IsObserving(r0))
	testutil.Equal(t, true, g.IsObserving(m0))
	testutil.Equal(t, true, g.IsObserving(m1))
	testutil.Equal(t, false, g.IsObserving(m2), "we observed m1, which is the parent of m2!")

	testutil.Equal(t, false, g.IsObserving(ar0))
	testutil.Equal(t, false, g.IsObserving(am0))
	testutil.Equal(t, false, g.IsObserving(am1))
	testutil.Equal(t, false, g.IsObserving(am2))

	g.unobserveNodes(ctx, am1, o)

	testutil.Equal(t, true, g.IsObserving(r0))
	testutil.Equal(t, true, g.IsObserving(m0))
	testutil.Equal(t, true, g.IsObserving(m1))
	testutil.Equal(t, false, g.IsObserving(m2))
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

	v := Var(Root(), "hello")
	o := Observe(Root(), g, v)
	_, ok := g.observers[o.Node().ID()]
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 2, g.numNodes)
	testutil.Equal(t, 1, o.Node().height)
	testutil.Equal(t, true, g.recomputeHeap.has(o))
	g.recomputeHeap.remove(o)
	testutil.Equal(t, false, g.recomputeHeap.has(o))

	g.addObserver(o)
	testutil.Equal(t, 2, g.numNodes)
	testutil.Equal(t, 1, o.Node().height)
	testutil.Equal(t, false, g.recomputeHeap.has(o))
}

func Test_Graph_recompute_recomputesObservers(t *testing.T) {
	g := New()
	n := newMockBareNode()
	o := Observe(Root(), g, n)
	g.recomputeHeap.Clear()

	testutil.Equal(t, false, g.recomputeHeap.has(n))
	testutil.Equal(t, false, g.recomputeHeap.has(o))

	err := g.recompute(testContext(), n)
	testutil.Nil(t, err)
	testutil.Equal(t, 1, g.recomputeHeap.len())
	testutil.Equal(t, true, g.recomputeHeap.has(o))
}
