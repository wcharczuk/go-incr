package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_adjustHeightsHeap_add(t *testing.T) {
	ah := newAdjustHeightsHeap(8)

	mn00 := newMockBareNodeWithHeight(0)
	mn10 := newMockBareNodeWithHeight(1)
	mn11 := newMockBareNodeWithHeight(1)
	mn20 := newMockBareNodeWithHeight(2)
	mn21 := newMockBareNodeWithHeight(2)
	mn22 := newMockBareNodeWithHeight(2)

	ah.add(mn00)
	ah.add(mn10)
	ah.add(mn11)
	ah.add(mn20)
	ah.add(mn21)
	ah.add(mn22)

	testutil.Equal(t, 0, ah.heightLowerBound)
	testutil.Equal(t, 2, ah.maxHeightSeen)
	testutil.Equal(t, 6, ah.len())
	testutil.Equal(t, 1, len(ah.nodesByHeight[0]))
	testutil.Equal(t, 2, len(ah.nodesByHeight[1]))
	testutil.Equal(t, 3, len(ah.nodesByHeight[2]))

	_, _ = ah.removeMinUnsafe()
	_, _ = ah.removeMinUnsafe()

	testutil.Equal(t, 1, ah.heightLowerBound)
	testutil.Equal(t, 2, ah.maxHeightSeen)
}
