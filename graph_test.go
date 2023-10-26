package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_New(t *testing.T) {
	r0 := Return("hello")
	r1 := Return("world!")
	m0 := Map2(r0, r1, func(v0, v1 string) string { return v0 + v1 })
	g := New()
	_ = Observe(g, m0)

	testutil.ItsEqual(t, true, g.IsObserving(r0))
	testutil.ItsEqual(t, true, g.IsObserving(r1))
	testutil.ItsEqual(t, true, g.IsObserving(m0))

	m1 := Map2(r0, r1, func(v0, v1 string) string { return v0 + v1 })
	testutil.ItsEqual(t, false, g.IsObserving(m1))
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

func Test_Graph_UndiscoverNodes(t *testing.T) {
	r0 := Return("hello")
	m0 := Map(r0, ident)
	m1 := Map(m0, ident)
	m2 := Map(m1, ident)

	ar0 := Return("hello")
	am0 := Map(ar0, ident)
	am1 := Map(am0, ident)
	am2 := Map(am1, ident)

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

	g.UndiscoverNodes(o1, m1)

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

func Test_Graph_UndiscoverNodes_notObserving(t *testing.T) {
	r0 := Return("hello")
	m0 := Map(r0, ident)
	m1 := Map(m0, ident)
	m2 := Map(m1, ident)

	ar0 := Return("hello")
	am0 := Map(ar0, ident)
	am1 := Map(am0, ident)
	am2 := Map(am1, ident)

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

	g.UndiscoverNodes(o, am1)

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

func Test_Graph_RecomputeHeight(t *testing.T) {
	g := New()

	n0 := emptyNode{NewNode()}
	n1 := emptyNode{NewNode()}
	n2 := emptyNode{NewNode()}
	n3 := emptyNode{NewNode()}

	Link(n1, n0)
	Link(n2, n1)
	Link(n3, n2)

	g.RecomputeHeight(n1)

	testutil.ItsEqual(t, 0, n0.n.height)
	testutil.ItsEqual(t, 2, n1.n.height)
	testutil.ItsEqual(t, 3, n2.n.height)
	testutil.ItsEqual(t, 4, n3.n.height)
}

func Test_Graph_RecomputeHeight_observed(t *testing.T) {
	g := New()

	v0 := Var("a")
	m0 := Map(v0, ident)
	o0 := Observe(g, m0)

	m1 := Map(m0, ident)
	m2 := Map(m1, ident)
	o1 := Observe(g, m2)

	m0.Node().height = 1
	g.RecomputeHeight(m0)

	_ = g.Stabilize(context.TODO())

	testutil.ItsEqual(t, "a", o0.Value())
	testutil.ItsEqual(t, "a", o1.Value())
}

func Test_Graph_DiscoverObserver_rediscover(t *testing.T) {
	g := New()

	v := Var("hello")
	o := Observe(g, v)
	_, ok := g.observers[o.Node().ID()]
	testutil.ItsEqual(t, true, ok)
	testutil.ItsEqual(t, 2, g.numNodes)
	testutil.ItsEqual(t, 2, o.Node().height)
	testutil.ItsEqual(t, true, g.recomputeHeap.Has(o))
	g.recomputeHeap.Remove(o)
	testutil.ItsEqual(t, false, g.recomputeHeap.Has(o))

	g.DiscoverObserver(o)
	testutil.ItsEqual(t, 2, g.numNodes)
	testutil.ItsEqual(t, 2, o.Node().height)
	testutil.ItsEqual(t, false, g.recomputeHeap.Has(o))
}
