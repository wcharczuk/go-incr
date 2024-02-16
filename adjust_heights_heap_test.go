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

func Test_adjustHeightsHeap_removeUnsafe(t *testing.T) {
	g := New()
	ahh := newAdjustHeightsHeap(32)

	n := newMockBareNode(g)
	ahh.removeUnsafe(n)

	n0 := newMockBareNodeWithHeight(g, 1)
	ahh.addUnsafe(n0)
	testutil.Equal(t, 1, n0.Node().heightInAdjustHeightsHeap)
	n1 := newMockBareNodeWithHeight(g, 2)
	ahh.addUnsafe(n1)
	testutil.Equal(t, 2, n1.Node().heightInAdjustHeightsHeap)

	ahh.removeUnsafe(n)

	testutil.Equal(t, 1, ahh.nodesByHeight[2].len())
	testutil.Equal(t, 2, ahh.numNodes)

	ahh.removeUnsafe(n1)
	testutil.Equal(t, 1, ahh.nodesByHeight[2].len())
	testutil.Equal(t, 2, ahh.numNodes)

	ahh.heightLowerBound = 0
	ahh.removeUnsafe(n1)
	testutil.Equal(t, 0, ahh.nodesByHeight[2].len())
	testutil.Equal(t, 1, ahh.numNodes)
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
