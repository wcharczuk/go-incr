package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_adjustHeightsHeap_len(t *testing.T) {
	g := New()
	ahh := newAdjustHeightsHeap(32)

	testutil.Equal(t, 0, ahh.len())

	n := newMockBareNode(g)
	ahh.addUnsafe(n)

	testutil.Equal(t, 1, ahh.len())
}

func Test_adjustHeightsHeap_maxHeightAllowed(t *testing.T) {
	ahh := newAdjustHeightsHeap(32)

	testutil.Equal(t, 31, ahh.maxHeightAllowed())
}

func Test_adjustHeightsHeap_setHeightUnsafe(t *testing.T) {
	g := New()
	ahh := newAdjustHeightsHeap(32)

	n0 := newMockBareNodeWithHeight(g, 1)
	ahh.maxHeightSeen = 10
	err := ahh.setHeightUnsafe(n0, 2)
	testutil.NoError(t, err)
	testutil.Equal(t, 2, n0.Node().height)
	testutil.Equal(t, 10, ahh.maxHeightSeen)

	n1 := newMockBareNodeWithHeight(g, 2)
	err = ahh.setHeightUnsafe(n1, ahh.maxHeightAllowed()+1)
	testutil.Error(t, err)
	testutil.Equal(t, 2, n0.Node().height)

	testutil.Equal(t, 10, ahh.maxHeightSeen)
	n2 := newMockBareNodeWithHeight(g, 1)
	err = ahh.setHeightUnsafe(n2, 11)
	testutil.NoError(t, err)
	testutil.Equal(t, 11, n2.Node().height)
	testutil.Equal(t, 11, ahh.maxHeightSeen)
}

func Test_adjustHeightsHeap_ensureHeightRequirementUnsafe(t *testing.T) {
	ahh := newAdjustHeightsHeap(32)

	g := New()
	n0 := newMockBareNodeWithHeight(g, 1)
	n1 := newMockBareNodeWithHeight(g, 2)
	n2 := newMockBareNodeWithHeight(g, 3)
	n3 := newMockBareNodeWithHeight(g, 4)

	err := ahh.ensureHeightRequirementUnsafe(n0, n1, n1, n0)
	testutil.Error(t, err)
	testutil.Equal(t, 0, ahh.numNodes)

	err = ahh.ensureHeightRequirementUnsafe(n0, n1, n2, n3)
	testutil.NoError(t, err)
	testutil.Equal(t, 1, ahh.numNodes)
	testutil.Equal(t, n3.Node().height+1, n2.Node().height)

	n4 := newMockBareNodeWithHeight(g, 5)
	n5 := newMockBareNodeWithHeight(g, ahh.maxHeightAllowed()+1)

	err = ahh.ensureHeightRequirementUnsafe(n0, n1, n4, n5)
	testutil.Error(t, err)
	testutil.Equal(t, 2, ahh.numNodes)
}

func Test_adjustHeightsHeap_adjustHeights(t *testing.T) {
	g := New()
	ahh := newAdjustHeightsHeap(32)

	n4 := newMockBareNodeWithHeight(g, 5)
	n5 := newMockBareNodeWithHeight(g, ahh.maxHeightAllowed()+1)

	err := ahh.adjustHeights(g.recomputeHeap, n4, n5)
	testutil.Error(t, err, "we should error on the original parent being beyond the maximum height")
	testutil.Equal(t, 5, ahh.heightLowerBound, "we should still set the height lower bound on error")
}
