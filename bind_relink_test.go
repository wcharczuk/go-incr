package incr

import (
	"context"
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

// edgeAsymmetries reports edges recorded on one side of a parent/child pair but
// not the other, counting multiplicity.
//
// Linking is symmetric -- [Graph.link] records the edge on both nodes -- so any
// asymmetry means some unlink path dropped only one side, leaving a reference
// that nothing will ever remove.
func edgeAsymmetries(g *Graph) []string {
	var problems []string
	count := func(list []INode, id Identifier) (n int) {
		for _, x := range list {
			if x.Node().id == id {
				n++
			}
		}
		return
	}
	for _, n := range g.nodes {
		nn := n.Node()
		for _, p := range nn.parents {
			if fwd, back := count(nn.parents, p.Node().id), count(p.Node().children, nn.id); fwd != back {
				problems = append(problems, fmt.Sprintf("%s lists parent %s x%d, but that node lists it as a child x%d",
					nn.kind, p.Node().kind, fwd, back))
			}
		}
		for _, c := range nn.children {
			if fwd, back := count(nn.children, c.Node().id), count(c.Node().parents, nn.id); fwd != back {
				problems = append(problems, fmt.Sprintf("%s lists child %s x%d, but that node lists it as a parent x%d",
					nn.kind, c.Node().kind, fwd, back))
			}
		}
	}
	return problems
}

// Test_changeParent_unlinksBothDirections covers a bind rewriting its
// right-hand side.
//
// Swapping the right-hand side has to unlink the old one from both ends. Removing
// the child from the old parent alone leaves the old parent in the child's parent
// list, where no later pass removes it: the list then grows by one entry per
// rebuild, holding every superseded subgraph reachable forever and lengthening
// the walks that staleness and invalidation checks make over it.
func Test_changeParent_unlinksBothDirections(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(256))

	sel := Var(g, 0)
	b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
		return Map(bs, Return(bs, which), func(x int) int { return x + 1 })
	})
	o := MustObserve(g, b)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.Equal(t, 1, o.Value())

	// a bind main takes exactly two inputs once it has a right-hand side: its
	// lhs-change node and the current right-hand side.
	const wantParents = 2
	for i := 1; i <= 50; i++ {
		sel.Set(i)
		testutil.Nil(t, g.Stabilize(ctx))
		testutil.Equal(t, i+1, o.Value())
		testutil.Equal(t, wantParents, len(b.Node().parents))
		testutil.Equal(t, 0, len(edgeAsymmetries(g)))
	}
}

// Test_changeParent_unlinksBothDirectionsNested covers the same invariant on a
// graph where many binds share one input, which is where a per-rebuild leak
// compounds fastest.
func Test_changeParent_unlinksBothDirectionsNested(t *testing.T) {
	ctx := context.Background()
	g := New(OptGraphMaxHeight(2048))

	control := Var(g, 1)
	o := makeNestedBindGraph(g, 6, control)

	testutil.Nil(t, g.Stabilize(ctx))
	testutil.NotEqual(t, 0, o.Value())

	for i := 1; i <= 8; i++ {
		control.Set(i)
		testutil.Nil(t, g.Stabilize(ctx))
		if problems := edgeAsymmetries(g); len(problems) > 0 {
			t.Fatalf("pass %d: %d edge asymmetries, first: %s", i, len(problems), problems[0])
		}
	}
}
