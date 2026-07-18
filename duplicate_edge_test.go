package incr

import (
	"context"
	"io"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

// This file exercises graphs where the same node appears more than once in
// another node's input list, which happens whenever a node takes one input in
// several slots -- Map2(x, x), Map3(x, x, x), MapIf(x, x, p).
//
// Every place the graph walks an edge list has to cope with that, because
// linking and unlinking work by identity: unlinking a pair drops every edge
// between them at once, and a node already queued somewhere must not be queued
// again. Three separate bugs in this family have been found (see
// _bench/ALGORITHMS.md), each in a walk whose author had not considered it, so
// these tests drive duplicate-input graphs through each walk deliberately rather
// than relying on the incidental guards that happen to make most of them safe.
//
// The assertions are the graph's structural invariants: computed values, an exact
// and stable node count, a self-consistent recompute heap, and a fully drained
// graph after unobserving.

// assertGraphInvariants checks the bookkeeping that the duplicate-edge bugs
// corrupted: heap self-consistency and an empty heap once stabilization settles.
func assertGraphInvariants(t *testing.T, g *Graph) {
	t.Helper()
	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, 0, g.recomputeHeap.len())
}

// Test_duplicateEdge_constructAndUpdate drives the construction walk
// (becameNecessaryRecursive/link) and the update propagation walk (the children
// loop in recompute) over nodes that take one input several times.
func Test_duplicateEdge_constructAndUpdate(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 1)
	two := Map2(g, v, v, func(a, b int) int { return a + b })
	three := Map3(g, two, two, two, func(a, b, c int) int { return a + b + c })
	o := MustObserve(g, three)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 6, o.Value())
	assertGraphInvariants(t, g)
	// var + map2 + map3 + observer
	testutil.Equal(t, 4, ExpertGraph(g).NumNodes())

	for i := 2; i <= 5; i++ {
		v.Set(i)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 6*i, o.Value())
		assertGraphInvariants(t, g)
		testutil.Equal(t, 4, ExpertGraph(g).NumNodes())
	}
}

// Test_duplicateEdge_heightAdjustment drives the adjust-heights walk, which
// iterates a node's children while raising heights. A bind is used because that
// is what forces heights to be repaired after construction.
func Test_duplicateEdge_heightAdjustment(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(256))

	sel := Var(g, 1)
	shared := Var(g, 10)
	// the bind's right-hand side takes `shared` in several slots, and the
	// subgraph's height depends on the bind, so linking it forces a height
	// adjustment across duplicate edges.
	b := Bind(g, sel, func(bs Scope, depth int) Incr[int] {
		var cur Incr[int] = shared
		for i := 0; i < depth; i++ {
			cur = Map3(bs, cur, cur, cur, func(a, b, c int) int { return a + b - c })
		}
		return cur
	})
	o := MustObserve(g, b)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 10, o.Value())
	assertGraphInvariants(t, g)

	// growing and shrinking the subgraph repeatedly re-runs height adjustment
	// over the duplicate edges in both directions.
	for _, depth := range []int{4, 1, 8, 2, 6} {
		sel.Set(depth)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, 10, o.Value())
		assertGraphInvariants(t, g)
	}

	// changing the shared input has to propagate through the duplicated edges.
	shared.Set(21)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 21, o.Value())
	assertGraphInvariants(t, g)
}

// Test_duplicateEdge_invalidation drives the invalidation walk, which pushes
// every child of an invalidated node onto a queue. A bind rebuild invalidates
// the nodes it created, so duplicate edges are walked while tearing down.
func Test_duplicateEdge_invalidation(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(256))

	sel := Var(g, 0)
	b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
		inner := Map(bs, Return(bs, which), func(x int) int { return x + 1 })
		// several nodes each taking `inner` more than once, so invalidating
		// `inner` reaches the same child through repeated edges.
		left := Map2(bs, inner, inner, func(a, b int) int { return a + b })
		right := Map3(bs, inner, inner, inner, func(a, b, c int) int { return a + b + c })
		return Map2(bs, left, right, func(a, b int) int { return a + b })
	})
	o := MustObserve(g, b)

	testutil.Nil(t, g.Stabilize(ctx))
	// inner = 1, left = 2, right = 3, total = 5
	testutil.Equal(t, 5, o.Value())
	assertGraphInvariants(t, g)
	// sel + bind main + bind lhs-change + observer + return + map + map2 + map3 + map2
	const wantNodes = 9
	testutil.Equal(t, wantNodes, ExpertGraph(g).NumNodes())

	for i := 1; i <= 4; i++ {
		sel.Set(i)
		testutil.Nil(t, g.Stabilize(ctx))
		inner := i + 1
		testutil.Equal(t, 2*inner+3*inner, o.Value())
		assertGraphInvariants(t, g)
		// the node count must return to the same place; drifting means a node was
		// registered or deregistered more than once.
		testutil.Equal(t, wantNodes, ExpertGraph(g).NumNodes())
	}
}

// Test_duplicateEdge_unobserve drives the teardown walk to completion: after
// unobserving, every node must be deregistered exactly once.
func Test_duplicateEdge_unobserve(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 3)
	m := Map(g, v, func(x int) int { return x * 2 })
	dup := Map3(g, m, m, m, func(a, b, c int) int { return a + b + c })
	mif := MapIf(g, dup, dup, Return(g, true))
	o := MustObserve(g, mif)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 18, o.Value())
	assertGraphInvariants(t, g)

	o.Unobserve(ctx)
	testutil.Equal(t, 0, ExpertGraph(g).NumNodes())
	testutil.Equal(t, 0, g.recomputeHeap.len())

	// re-observing has to rebuild cleanly from a fully torn down graph.
	o2 := MustObserve(g, mif)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 18, o2.Value())
	assertGraphInvariants(t, g)
}

// Test_duplicateEdge_sentinel drives the sentinel walk over a node whose inputs
// repeat.
func Test_duplicateEdge_sentinel(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 1)
	// recomputes counts how often the duplicate-input node actually recomputes,
	// which is what the sentinel drives.
	var recomputes int
	dup := Map2(g, v, v, func(a, b int) int {
		recomputes++
		return a + b
	})
	// the sentinel keeps the node recomputing for the first few passes; a
	// sentinel does not make its watched node necessary, so it is only reached
	// through the graph's sentinel walk.
	s := Sentinel(g, func() bool { return recomputes < 3 }, dup)
	o := MustObserve(g, dup)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, o.Value())
	testutil.Equal(t, 1, recomputes)
	testutil.Nil(t, g.recomputeHeap.sanityCheck())

	// while the predicate holds, the sentinel re-stales the node each pass.
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, recomputes)
	testutil.Nil(t, g.recomputeHeap.sanityCheck())

	s.Unwatch(ctx)
	v.Set(5)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 10, o.Value())
	assertGraphInvariants(t, g)
}

// Test_duplicateEdge_parallelStabilize drives the parallel stabilization path,
// which walks children under its own locking and uses a different iterator over
// the recompute heap than the serial path.
func Test_duplicateEdge_parallelStabilize(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(256))

	v := Var(g, 1)
	dup := Map3(g, v, v, v, func(a, b, c int) int { return a + b + c })
	fanout := make([]Incr[int], 8)
	for i := range fanout {
		fanout[i] = Map2(g, dup, dup, func(a, b int) int { return a + b })
	}
	root := fanout[0]
	for _, f := range fanout[1:] {
		root = Map2(g, root, f, func(a, b int) int { return a + b })
	}
	o := MustObserve(g, root)

	testutil.Nil(t, g.ParallelStabilize(ctx))
	testutil.Equal(t, 8*6, o.Value())
	assertGraphInvariants(t, g)

	for i := 2; i <= 4; i++ {
		v.Set(i)
		testutil.Nil(t, g.ParallelStabilize(ctx))
		testutil.Equal(t, 8*6*i, o.Value())
		assertGraphInvariants(t, g)
	}
}

// Test_duplicateEdge_dotAndCycleDetection drives the two remaining walks: the
// diagram renderer, and the cycle check that follows edges looking for a path
// back to a node.
func Test_duplicateEdge_dotAndCycleDetection(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 1)
	dup := Map3(g, v, v, v, func(a, b, c int) int { return a + b + c })
	o := MustObserve(g, dup)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 3, o.Value())

	// rendering must terminate and not fault on repeated edges.
	testutil.Nil(t, Dot(io.Discard, g))

	// linking dup back into its own input would close a cycle; the check has to
	// find it through the duplicated edges rather than looping.
	testutil.NotNil(t, DetectCycleIfLinked(v, dup))
	// and an unrelated pair must still be reported as safe to link.
	other := Var(g, 2)
	testutil.Nil(t, DetectCycleIfLinked(dup, other))
}
