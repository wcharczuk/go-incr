package incr

import (
	"context"
	"strings"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

// Test_CheckInvariants_cleanGraph checks that an ordinary graph, built and stabilized the
// usual way, reports nothing.
func Test_CheckInvariants_cleanGraph(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(64))

	v := Var(g, 1)
	m := Map(g, v, func(x int) int { return x + 1 })
	sel := Var(g, 0)
	b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
		return Map2(bs, m, Return(bs, which), func(a, c int) int { return a + c })
	})
	MustObserve(g, b)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())

	// including after a bind has rewritten its subgraph a few times
	for i := 1; i <= 3; i++ {
		sel.Set(i)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Nil(t, ExpertGraph(g).CheckInvariants())
	}
}

// Test_CheckInvariants_catchesHalfEdge covers what the check is for: a caller using the
// one-sided edge methods and pairing them up wrongly.
//
// Those methods change one node's record of an edge and leave the other alone, which is
// what a caller rebuilding a graph from outside needs. The cost is that a mistake is
// quiet, because nothing will ever remove the dangling half. This is the assertion that
// such a mistake is at least detectable.
func Test_CheckInvariants_catchesHalfEdge(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(64))

	v := Var(g, 1)
	m := Map(g, v, func(x int) int { return x + 1 })
	MustObserve(g, m)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())

	// record the edge on one side only, as a caller pairing these up wrongly would
	other := Map(g, v, func(x int) int { return x })
	MustObserve(g, other)
	testutil.Nil(t, g.Stabilize(ctx))
	ExpertNode(m).AddParents(other)

	err := ExpertGraph(g).CheckInvariants()
	testutil.NotNil(t, err, "a half-recorded edge should be reported")
	if !strings.Contains(err.Error(), "edge asymmetry") {
		t.Fatalf("expected an edge asymmetry, got: %v", err)
	}

	// Pairing up the edge is necessary but not sufficient: other and m are siblings at
	// the same height, so making one the parent of the other leaves a height inversion.
	// A caller wiring edges by hand owns both invariants, which is what the check is for.
	ExpertNode(other).AddChildren(m)
	err = ExpertGraph(g).CheckInvariants()
	testutil.NotNil(t, err, "the edge now agrees, but the heights do not")
	if !strings.Contains(err.Error(), "height inversion") {
		t.Fatalf("expected a height inversion, got: %v", err)
	}

	// with both halves recorded and the height raised above the new parent, the graph is
	// consistent again
	ExpertNode(m).SetHeight(ExpertNode(other).Height() + 1)
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())
}

// Test_CheckInvariants_catchesHeightInversion covers the other invariant a caller takes
// on: a child has to sit above its parents, or a pass can compute it from a stale input.
func Test_CheckInvariants_catchesHeightInversion(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(64))

	v := Var(g, 1)
	m := Map(g, v, func(x int) int { return x + 1 })
	MustObserve(g, m)
	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Nil(t, ExpertGraph(g).CheckInvariants())

	// SetHeight is exposed, so a caller restoring heights from a store can invert them
	ExpertNode(m).SetHeight(0)

	err := ExpertGraph(g).CheckInvariants()
	testutil.NotNil(t, err, "an inverted height should be reported")
	if !strings.Contains(err.Error(), "height inversion") {
		t.Fatalf("expected a height inversion, got: %v", err)
	}
}

// Test_CheckInvariants_reportsEveryProblem checks that the whole graph is reported rather
// than the first thing found, since one bad assembly step tends to produce several.
func Test_CheckInvariants_reportsEveryProblem(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(64))

	v := Var(g, 1)
	first := Map(g, v, func(x int) int { return x })
	second := Map(g, v, func(x int) int { return x })
	MustObserve(g, first)
	MustObserve(g, second)
	testutil.Nil(t, g.Stabilize(ctx))

	ExpertNode(first).AddParents(second)
	ExpertNode(second).AddParents(first)

	err := ExpertGraph(g).CheckInvariants()
	testutil.NotNil(t, err)
	if got := strings.Count(err.Error(), "edge asymmetry"); got < 2 {
		t.Fatalf("expected both problems reported, got %d in: %v", got, err)
	}
}
