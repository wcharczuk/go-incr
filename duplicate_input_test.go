package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

// Test_removeParents_duplicateInput covers tearing down a node that takes the
// same input in more than one slot.
//
// Unlinking removes every edge between a pair at once, so the duplicate entries
// in the input list must not each trigger a teardown of the parent. When they
// did, the parent's whole subgraph was re-walked once per duplicate edge, and
// the graph's node count was decremented once per repeat.
func Test_removeParents_duplicateInput(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 1)
	sel := Var(g, 0)
	// the bind's subgraph feeds one node into all three slots of a map3, so the
	// map3's input list holds that node three times.
	b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
		inner := Map(bs, v, func(x int) int { return x + which })
		return Map3(bs, inner, inner, inner, func(a, b, c int) int { return a + b + c })
	})
	o := MustObserve(g, b)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 3, o.Value())

	// var + sel + bind main + bind lhs-change + observer + inner map + map3
	const wantNodes = 7
	testutil.Equal(t, wantNodes, ExpertGraph(g).NumNodes())

	// each rebuild discards the previous subgraph and builds a new one; the node
	// count has to come back to the same place every time rather than drifting.
	for i := 1; i <= 5; i++ {
		sel.Set(i)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 3*(1+i), o.Value())
		testutil.Equal(t, wantNodes, ExpertGraph(g).NumNodes())
	}
}

// Test_removeParents_duplicateInputIsLinear covers the cost, not just the
// bookkeeping: a chain of nodes that each take the same input several times used
// to cost arity^depth to tear down, because every duplicate edge re-walked the
// whole subgraph below it.
//
// Node counts are asserted rather than timings: if the teardown were still
// re-walking, the repeated removals would show up here as a wrong count, and at
// this depth the test would take minutes rather than milliseconds.
func Test_removeParents_duplicateInputIsLinear(t *testing.T) {
	ctx := context.Background()
	const depth = 64
	g := New(OptGraphMaxHeight(8 * depth))

	sel := Var(g, 0)
	b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
		var cur Incr[int] = Return(bs, which)
		for i := 0; i < depth; i++ {
			cur = Map3(bs, cur, cur, cur, func(a, b, c int) int { return a + b - c })
		}
		return cur
	})
	o := MustObserve(g, b)

	testutil.Nil(t, g.Stabilize(ctx))
	// sel + bind main + bind lhs-change + observer + return + depth map3 nodes
	wantNodes := 5 + depth
	testutil.Equal(t, wantNodes, ExpertGraph(g).NumNodes())

	for i := 1; i <= 3; i++ {
		sel.Set(i)
		testutil.Nil(t, g.Stabilize(ctx))
		// map3(x,x,x) = x + x - x = x, so the chain is the identity on the seed.
		testutil.Equal(t, i, o.Value())
		testutil.Equal(t, wantNodes, ExpertGraph(g).NumNodes())
	}
}

// Test_removeParents_duplicateInputUnobserve covers the same duplicate-edge
// teardown reached by unobserving rather than by a bind rebuild.
func Test_removeParents_duplicateInputUnobserve(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 2)
	m := Map(g, v, func(x int) int { return x * 2 })
	same := Map3(g, m, m, m, func(a, b, c int) int { return a + b + c })
	o := MustObserve(g, same)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 12, o.Value())
	testutil.Equal(t, 4, ExpertGraph(g).NumNodes())

	o.Unobserve(ctx)
	// unobserving the only observer makes the whole graph unnecessary; every node
	// should be deregistered exactly once.
	testutil.Equal(t, 0, ExpertGraph(g).NumNodes())
	testutil.Equal(t, 0, g.recomputeHeap.len())
}
