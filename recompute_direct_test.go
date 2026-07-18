package incr

import (
	"context"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

// Test_recompute_directly_duplicateChild covers a node that appears twice in its
// parent's children list, which happens whenever a node takes the same input
// more than once.
//
// A child held back for direct recomputation must not also be queued in the
// recompute heap. If it is, it gets recomputed while still linked into a height
// block, and the heap's bookkeeping no longer matches its lists.
func Test_recompute_directly_duplicateChild(t *testing.T) {
	ctx := context.Background()
	g := New()

	v := Var(g, 1)
	// both inputs are the same node, so v's children contain this map2 twice.
	same := Map2(g, v, v, func(a, b int) int { return a + b })
	o := MustObserve(g, same)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 2, o.Value())
	testutil.Nil(t, g.recomputeHeap.sanityCheck())

	v.Set(5)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 10, o.Value())
	testutil.Nil(t, g.recomputeHeap.sanityCheck())
	testutil.Equal(t, 0, g.recomputeHeap.len())
}

// Test_recompute_directly_chainsWholeChain asserts the direct-recompute path
// actually engages for a chain of single-input nodes, since a regression would
// be silent: the graph would still compute correct values, just by way of the
// recompute heap for every node.
func Test_recompute_directly_chainsWholeChain(t *testing.T) {
	ctx := context.Background()
	const depth = 32
	g := New(OptGraphMaxHeight(depth + 16))

	v := Var(g, 0)
	var cur Incr[int] = v
	for i := 0; i < depth; i++ {
		cur = Map(g, cur, func(x int) int { return x + 1 })
	}
	o := MustObserve(g, cur)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, depth, o.Value())

	before := g.numNodesRecomputedDirectly
	v.Set(10)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, depth+10, o.Value())

	// every node past the var should have been reached by chaining rather than
	// through the recompute heap.
	testutil.Equal(t, depth, int(g.numNodesRecomputedDirectly-before))
}

// Test_recompute_directly_multiInputWaits asserts a node with more than one input
// is not recomputed directly, because a sibling input may still be pending.
func Test_recompute_directly_multiInputWaits(t *testing.T) {
	ctx := context.Background()
	g := New()

	a := Var(g, 1)
	b := Var(g, 2)
	// deepen one side so the two inputs to the sum sit at different heights.
	deeper := Map(g, Map(g, a, func(x int) int { return x }), func(x int) int { return x })
	sum := Map2(g, deeper, b, func(x, y int) int { return x + y })
	o := MustObserve(g, sum)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 3, o.Value())

	before := g.numNodesRecomputedDirectly
	a.Set(10)
	b.Set(20)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 30, o.Value())

	// the two single-input maps above `a` may chain, but the map2 must not, so
	// the count cannot reach every recomputed node.
	testutil.Equal(t, true, int(g.numNodesRecomputedDirectly-before) <= 2)
}
