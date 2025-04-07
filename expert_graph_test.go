package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ExpertGraph_SetID(t *testing.T) {
	g := New()
	eg := ExpertGraph(g)

	newID := _defaultIdentifierProvider.NewIdentifier()
	eg.SetID(newID)
	testutil.Equal(t, newID, g.ID())
}

func Test_ExpertGraph_SetStabilizationNum(t *testing.T) {
	g := New()
	eg := ExpertGraph(g)
	eg.SetStabilizationNum(1234)
	testutil.Equal(t, 1234, eg.StabilizationNum())
}

func Test_ExpertGraph_RecomputeHeapAdd(t *testing.T) {
	g := New()
	eg := ExpertGraph(g)

	n1 := newMockBareNode(g)
	n2 := newMockBareNode(g)

	eg.RecomputeHeapAdd(n1, n2)
	testutil.Equal(t, 2, g.recomputeHeap.len())
	testutil.Equal(t, 2, eg.RecomputeHeapLen())
}

func Test_ExpertGraph_SetDuringStabilizationAdd(t *testing.T) {
	g := New()
	eg := ExpertGraph(g)

	n1 := newMockBareNode(g)
	n2 := newMockBareNode(g)

	eg.SetDuringStabilizationAdd(n1, n2)
	testutil.Equal(t, 2, len(g.setDuringStabilization))
	testutil.Equal(t, 2, len(eg.SetDuringStabilizationIDs()))
}

func Test_ExpertGraph_stats(t *testing.T) {
	g := New()

	g.numNodes = 12
	g.numNodesChanged = 13
	g.numNodesRecomputed = 14
	eg := ExpertGraph(g)
	testutil.Equal(t, 12, eg.NumNodes())
	testutil.Equal(t, 13, eg.NumNodesChanged())
	testutil.Equal(t, 14, eg.NumNodesRecomputed())
}

func Test_ExpertGraph_RecomputeHeapIDs(t *testing.T) {
	g := New()
	eg := ExpertGraph(g)

	n1 := newMockBareNode(g)
	n2 := newMockBareNode(g)
	n2.n.height = 3

	eg.RecomputeHeapAdd(n1, n2)
	testutil.Equal(t, 2, g.recomputeHeap.len())

	recomputeHeapIDs := eg.RecomputeHeapIDs()
	testutil.Equal(t, 2, len(recomputeHeapIDs))
	testutil.Any(t, recomputeHeapIDs, func(id Identifier) bool { return id == n1.n.id })
	testutil.Any(t, recomputeHeapIDs, func(id Identifier) bool { return id == n2.n.id })
}
