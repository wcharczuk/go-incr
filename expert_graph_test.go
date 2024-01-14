package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ExpertGraph_SetID(t *testing.T) {
	g := New()
	eg := ExpertGraph(g)

	newID := NewIdentifier()
	eg.SetID(newID)

	testutil.ItsEqual(t, newID, g.ID())
}

func Test_ExpertGraph_SetStabilizationNum(t *testing.T) {
	g := New()
	eg := ExpertGraph(g)
	eg.SetStabilizationNum(1234)
	testutil.ItsEqual(t, 1234, eg.StabilizationNum())
}

func Test_ExpertGraph_RecomputeHeapAdd(t *testing.T) {
	g := New()
	eg := ExpertGraph(g)

	n1 := newMockBareNode()
	n2 := newMockBareNode()

	eg.RecomputeHeapAdd(n1, n2)
	testutil.ItsEqual(t, 2, g.recomputeHeap.Len())
	testutil.ItsEqual(t, 2, eg.RecomputeHeapLen())
}

func Test_ExpertGraph_stats(t *testing.T) {
	g := New()

	g.numNodes = 12
	g.numNodesChanged = 13
	g.numNodesRecomputed = 14
	eg := ExpertGraph(g)
	testutil.ItsEqual(t, 12, eg.NumNodes())
	testutil.ItsEqual(t, 13, eg.NumNodesChanged())
	testutil.ItsEqual(t, 14, eg.NumNodesRecomputed())
}

func Test_ExpertGraph_RecomputeHeap(t *testing.T) {
	g := New()
	eg := ExpertGraph(g)

	n1 := newMockBareNode()
	n2 := newMockBareNode()
	n2.n.height = 3

	eg.RecomputeHeapAdd(n1, n2)
	testutil.ItsEqual(t, 2, g.recomputeHeap.Len())

	recomputeHeapIDs := eg.RecomputeHeap()
	testutil.ItsEqual(t, 2, len(recomputeHeapIDs))
	testutil.ItsAny(t, recomputeHeapIDs, func(id Identifier) bool { return id == n1.n.id })
	testutil.ItsAny(t, recomputeHeapIDs, func(id Identifier) bool { return id == n2.n.id })
}
