package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_New(t *testing.T) {
	r0 := Return("hello")
	r1 := Return("world!")
	m0 := Map2(r0, r1, func(v0, v1 string) string { return v0 + v1 })
	g := New(m0)

	testutil.ItsEqual(t, true, g.IsObserving(r0))
	testutil.ItsEqual(t, true, g.IsObserving(r1))
	testutil.ItsEqual(t, true, g.IsObserving(m0))

	m1 := Map2(r0, r1, func(v0, v1 string) string { return v0 + v1 })
	testutil.ItsEqual(t, false, g.IsObserving(m1))
}

func Test_Graph_UndiscoverAllNodes(t *testing.T) {
	r0 := Return("hello")
	m0 := Map(r0, ident)
	m1 := Map(m0, ident)
	m2 := Map(m1, ident)

	ar0 := Return("hello")
	am0 := Map(ar0, ident)
	am1 := Map(am0, ident)
	am2 := Map(am1, ident)

	g := New()
	g.Observe(m1)
	g.Observe(am2)

	testutil.ItsEqual(t, true, g.IsObserving(r0))
	testutil.ItsEqual(t, true, g.IsObserving(m0))
	testutil.ItsEqual(t, true, g.IsObserving(m1))
	testutil.ItsEqual(t, true, g.IsObserving(m2))

	testutil.ItsEqual(t, true, g.IsObserving(ar0))
	testutil.ItsEqual(t, true, g.IsObserving(am0))
	testutil.ItsEqual(t, true, g.IsObserving(am1))
	testutil.ItsEqual(t, true, g.IsObserving(am2))

	g.UndiscoverAllNodes(m1)

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

func Test_Graph_undiscoverAllNodes_notObserving(t *testing.T) {
	r0 := Return("hello")
	m0 := Map(r0, ident)
	m1 := Map(m0, ident)
	m2 := Map(m1, ident)

	ar0 := Return("hello")
	am0 := Map(ar0, ident)
	am1 := Map(am0, ident)
	am2 := Map(am1, ident)

	g := New()
	g.Observe(m1)

	testutil.ItsEqual(t, true, g.IsObserving(r0))
	testutil.ItsEqual(t, true, g.IsObserving(m0))
	testutil.ItsEqual(t, true, g.IsObserving(m1))
	testutil.ItsEqual(t, true, g.IsObserving(m2))

	testutil.ItsEqual(t, false, g.IsObserving(ar0))
	testutil.ItsEqual(t, false, g.IsObserving(am0))
	testutil.ItsEqual(t, false, g.IsObserving(am1))
	testutil.ItsEqual(t, false, g.IsObserving(am2))

	g.UndiscoverAllNodes(am1)

	testutil.ItsEqual(t, true, g.IsObserving(r0))
	testutil.ItsEqual(t, true, g.IsObserving(m0))
	testutil.ItsEqual(t, true, g.IsObserving(m1))
	testutil.ItsEqual(t, true, g.IsObserving(m2))
}

func Test_Graph_IsStabilizing(t *testing.T) {
	g := New()
	testutil.ItsEqual(t, false, g.IsStabilizing())
	g.status = StatusStabilizing
	testutil.ItsEqual(t, true, g.IsStabilizing())
	g.status = StatusNotStabilizing
	testutil.ItsEqual(t, false, g.IsStabilizing())
}
