package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_recomputeHeap_add(t *testing.T) {
	g := New()

	rh := newRecomputeHeap(32)

	n50 := newHeightIncr(g, 5)
	n60 := newHeightIncr(g, 6)
	n70 := newHeightIncr(g, 7)

	rh.add(n50)

	// assertions post add n50
	{
		testutil.Equal(t, 1, rh.len())
		testutil.Equal(t, 1, rh.heights[5].len())
		testutil.Equal(t, true, rh.has(n50))
		testutil.Equal(t, false, rh.has(n60))
		testutil.Equal(t, false, rh.has(n70))
		testutil.Equal(t, 5, rh.minHeight)
		testutil.Equal(t, 5, rh.maxHeight)
	}

	rh.add(n60)

	// assertions post add n60
	{
		testutil.Equal(t, 2, rh.len())
		testutil.Equal(t, 1, rh.heights[5].len())
		testutil.Equal(t, 1, rh.heights[6].len())
		testutil.Equal(t, true, rh.has(n50))
		testutil.Equal(t, true, rh.has(n50))
		testutil.Equal(t, false, rh.has(n70))
		testutil.Equal(t, 5, rh.minHeight)
		testutil.Equal(t, 6, rh.maxHeight)
	}

	rh.add(n70)

	// assertions post add n70
	{
		testutil.Equal(t, 3, rh.len())
		testutil.Equal(t, 1, rh.heights[5].len())
		testutil.Equal(t, 1, rh.heights[6].len())
		testutil.Equal(t, 1, rh.heights[7].len())
		testutil.Equal(t, true, rh.has(n50))
		testutil.Equal(t, true, rh.has(n60))
		testutil.Equal(t, true, rh.has(n70))
		testutil.Equal(t, 5, rh.minHeight)
		testutil.Equal(t, 7, rh.maxHeight)
	}
}

func iterToArray[A any](fn func() (A, bool)) (output []A) {
	value, ok := fn()
	for ok {
		output = append(output, value)
		value, ok = fn()
	}
	return output
}

func Test_recomputeHeap_setIterToMinHeight(t *testing.T) {
	g := New()

	rh := newRecomputeHeap(10)

	n00 := newHeightIncr(g, 0)
	n01 := newHeightIncr(g, 0)
	n02 := newHeightIncr(g, 0)

	n10 := newHeightIncr(g, 1)
	n11 := newHeightIncr(g, 1)
	n12 := newHeightIncr(g, 1)
	n13 := newHeightIncr(g, 1)

	n50 := newHeightIncr(g, 5)
	n51 := newHeightIncr(g, 5)
	n52 := newHeightIncr(g, 5)
	n53 := newHeightIncr(g, 5)
	n54 := newHeightIncr(g, 5)

	rh.add(n00)
	rh.add(n01)
	rh.add(n02)
	rh.add(n10)
	rh.add(n11)
	rh.add(n12)
	rh.add(n13)
	rh.add(n50)
	rh.add(n51)
	rh.add(n52)
	rh.add(n53)
	rh.add(n54)

	testutil.Equal(t, 12, rh.len())
	testutil.Equal(t, 0, rh.minHeight)
	testutil.Equal(t, 5, rh.maxHeight)

	var iter recomputeHeapListIter
	var iterValues []INode
	rh.setIterToMinHeight(&iter)
	iterValues = iterToArray(iter.Next)

	testutil.Nil(t, rh.sanityCheck())
	testutil.Equal(t, 9, rh.len())
	testutil.Equal(t, 3, len(iterValues))
	testutil.Equal(t, 0, rh.heights[0].len())
	testutil.Equal(t, 4, rh.heights[1].len())
	testutil.Equal(t, 5, rh.heights[5].len())

	for _, n := range iterValues {
		testutil.NotNil(t, n)
		testutil.Nil(t, n.Node().nextInRecomputeHeap)
		testutil.Nil(t, n.Node().previousInRecomputeHeap)
	}

	testutil.Equal(t, 1, rh.minHeight)
	testutil.Equal(t, 5, rh.maxHeight)

	rh.setIterToMinHeight(&iter)
	iterValues = iterToArray(iter.Next)
	testutil.Nil(t, rh.sanityCheck())
	testutil.Equal(t, 5, rh.len())
	testutil.Equal(t, 4, len(iterValues))
	testutil.Equal(t, 0, rh.heights[0].len())
	testutil.Equal(t, 0, rh.heights[1].len())
	testutil.Equal(t, 5, rh.heights[5].len())

	for _, n := range iterValues {
		testutil.NotNil(t, n)
		testutil.Nil(t, n.Node().nextInRecomputeHeap)
		testutil.Nil(t, n.Node().previousInRecomputeHeap)
	}

	rh.setIterToMinHeight(&iter)
	iterValues = iterToArray(iter.Next)
	testutil.Nil(t, rh.sanityCheck())
	testutil.Equal(t, 0, rh.len())
	testutil.Equal(t, 5, len(iterValues))
	testutil.Equal(t, 0, rh.heights[0].len())
	testutil.Equal(t, 0, rh.heights[1].len())
	testutil.Equal(t, 0, rh.heights[5].len())

	for _, n := range iterValues {
		testutil.NotNil(t, n)
		testutil.Nil(t, n.Node().nextInRecomputeHeap)
		testutil.Nil(t, n.Node().previousInRecomputeHeap)
	}

	rh.add(n50)
	rh.add(n51)
	rh.add(n52)
	rh.add(n53)
	rh.add(n54)

	testutil.Nil(t, rh.sanityCheck())
	testutil.Equal(t, 5, rh.len())
	testutil.Equal(t, 5, rh.minHeight)
	testutil.Equal(t, 5, rh.maxHeight)

	rh.setIterToMinHeight(&iter)
	iterValues = iterToArray(iter.Next)
	testutil.Nil(t, rh.sanityCheck())
	testutil.Equal(t, 0, rh.len())
	testutil.Equal(t, 5, len(iterValues))
	testutil.Equal(t, 0, rh.heights[0].len())
	testutil.Equal(t, 0, rh.heights[1].len())
	testutil.Equal(t, 0, rh.heights[5].len())

	for _, n := range iterValues {
		testutil.NotNil(t, n)
		testutil.Nil(t, n.Node().nextInRecomputeHeap)
		testutil.Nil(t, n.Node().previousInRecomputeHeap)
	}
}

func Test_recomputeHeap_remove(t *testing.T) {
	g := New()
	rh := newRecomputeHeap(10)
	n10 := newHeightIncr(g, 1)
	n11 := newHeightIncr(g, 1)
	n20 := newHeightIncr(g, 2)
	n21 := newHeightIncr(g, 2)
	n22 := newHeightIncr(g, 2)
	n30 := newHeightIncr(g, 3)

	rh.add(n10)
	rh.add(n11)
	rh.add(n20)
	rh.add(n21)
	rh.add(n22)
	rh.add(n30)

	testutil.Nil(t, rh.sanityCheck())
	testutil.Equal(t, true, rh.has(n10))
	testutil.Equal(t, true, rh.has(n11))
	testutil.Equal(t, true, rh.has(n20))
	testutil.Equal(t, true, rh.has(n21))
	testutil.Equal(t, true, rh.has(n22))
	testutil.Equal(t, true, rh.has(n30))

	rh.remove(n21)

	testutil.Nil(t, rh.sanityCheck())
	testutil.Equal(t, 5, rh.len())
	testutil.Equal(t, true, rh.has(n10))
	testutil.Equal(t, true, rh.has(n11))
	testutil.Equal(t, true, rh.has(n20))
	testutil.Equal(t, false, rh.has(n21))

	for _, h := range rh.heights {
		if h == nil {
			continue
		}
		testutil.Equal(t, false, h.has(n21.n.id))
	}

	testutil.Equal(t, true, rh.has(n22))
	testutil.Equal(t, true, rh.has(n30))

	testutil.Equal(t, 2, rh.heights[2].len())

	rh.remove(n10)
	rh.remove(n11)

	testutil.Nil(t, rh.sanityCheck())
	testutil.Equal(t, 3, rh.len())
	testutil.Equal(t, 0, rh.heights[1].len())
	testutil.Equal(t, 2, rh.minHeight)
	testutil.Equal(t, 3, rh.maxHeight)

	for _, h := range rh.heights {
		if h == nil {
			continue
		}
		testutil.Equal(t, false, h.has(n10.n.id))
	}
	for _, h := range rh.heights {
		if h == nil {
			continue
		}
		testutil.Equal(t, false, h.has(n11.n.id))
	}
}

func Test_recomputeHeap_nextMinHeightUnsafe_noItems(t *testing.T) {
	rh := new(recomputeHeap)

	rh.minHeight = 1
	rh.maxHeight = 3

	next := rh.nextMinHeightUnsafe()
	testutil.Equal(t, 0, next)
}

func Test_recomputeHeap_maybeAddNewHeightsUnsafe(t *testing.T) {
	rh := newRecomputeHeap(8)
	testutil.Equal(t, 8, len(rh.heights))
	rh.maybeAddNewHeightsUnsafe(9) // we use (1) indexing!
	testutil.Equal(t, 10, len(rh.heights))
}

func Test_recomputeHeap_add_adjustsHeights(t *testing.T) {
	g := New()
	rh := newRecomputeHeap(8)
	testutil.Equal(t, 8, len(rh.heights))

	v0 := newHeightIncr(g, 32)
	rh.add(v0)
	testutil.Equal(t, 33, len(rh.heights))
	testutil.Equal(t, 32, rh.minHeight)
	testutil.Equal(t, 32, rh.maxHeight)

	v1 := newHeightIncr(g, 64)
	rh.add(v1)
	testutil.Equal(t, 65, len(rh.heights))
	testutil.Equal(t, 32, rh.minHeight)
	testutil.Equal(t, 64, rh.maxHeight)
}

func Test_recomputeHeap_fix(t *testing.T) {
	g := New()

	rh := newRecomputeHeap(8)
	v0 := newHeightIncr(g, 2)
	rh.add(v0)
	v1 := newHeightIncr(g, 3)
	rh.add(v1)
	v2 := newHeightIncr(g, 4)
	rh.add(v2)

	testutil.Equal(t, 2, rh.minHeight)
	testutil.Equal(t, 1, rh.heights[2].len())
	testutil.Equal(t, 1, rh.heights[3].len())
	testutil.Equal(t, 1, rh.heights[4].len())
	testutil.Equal(t, 4, rh.maxHeight)

	v0.n.height = 1
	rh.fix(v0)

	testutil.Equal(t, 1, rh.minHeight)
	testutil.Equal(t, 1, rh.heights[1].len())
	testutil.Equal(t, 0, rh.heights[2].len())
	testutil.Equal(t, 1, rh.heights[3].len())
	testutil.Equal(t, 1, rh.heights[4].len())
	testutil.Equal(t, 4, rh.maxHeight)

	rh.fix(v0)
	testutil.Equal(t, 1, rh.minHeight)
	testutil.Equal(t, 1, rh.heights[1].len())
	testutil.Equal(t, 0, rh.heights[2].len())
	testutil.Equal(t, 1, rh.heights[3].len())
	testutil.Equal(t, 1, rh.heights[4].len())
	testutil.Equal(t, 4, rh.maxHeight)

	v2.n.height = 5
	rh.fix(v2)

	testutil.Equal(t, 1, rh.minHeight)
	testutil.Equal(t, 1, rh.heights[1].len())
	testutil.Equal(t, 0, rh.heights[2].len())
	testutil.Equal(t, 1, rh.heights[3].len())
	testutil.Equal(t, 0, rh.heights[4].len())
	testutil.Equal(t, 1, rh.heights[5].len())
	testutil.Equal(t, 5, rh.maxHeight)
}

func Test_recomputeHeap_sanityCheck_badMinHeight(t *testing.T) {
	g := New()
	rh := newRecomputeHeap(8)

	n_2_00 := newMockBareNodeWithHeight(g, 2)
	n_2_01 := newMockBareNodeWithHeight(g, 2)

	rh.add(n_2_00)
	rh.add(n_2_01)

	rh.minHeight = 1

	err := rh.sanityCheck()
	testutil.NotNil(t, err)
}

func Test_recomputeHeap_sanityCheck_badItemHeight(t *testing.T) {
	g := New()
	rh := newRecomputeHeap(8)

	n_1_00 := newMockBareNodeWithHeight(g, 1)
	n_2_00 := newMockBareNodeWithHeight(g, 2)
	n_2_01 := newMockBareNodeWithHeight(g, 2)
	n_3_00 := newMockBareNodeWithHeight(g, 3)
	n_3_01 := newMockBareNodeWithHeight(g, 3)
	n_3_02 := newMockBareNodeWithHeight(g, 3)

	height2 := newList(n_2_00, n_2_01)

	rh.heights = []*recomputeHeapList{
		nil,
		newList(n_1_00),
		height2,
		newList(n_3_00, n_3_01, n_3_02),
	}

	n_2_00.Node().heightInRecomputeHeap = 1
	err := rh.sanityCheck()
	testutil.NotNil(t, err)
}

func Test_recomputeHeap_sanityCheck_badHeightInRecomputeHeap(t *testing.T) {
	g := New()
	rh := newRecomputeHeap(8)

	n_1_00 := newMockBareNodeWithHeight(g, 1)
	n_2_00 := newMockBareNodeWithHeight(g, 2)
	n_2_01 := newMockBareNodeWithHeight(g, 2)
	n_3_00 := newMockBareNodeWithHeight(g, 3)
	n_3_01 := newMockBareNodeWithHeight(g, 3)
	n_3_02 := newMockBareNodeWithHeight(g, 3)

	height2 := newList(n_2_00, n_2_01)

	rh.heights = []*recomputeHeapList{
		nil,
		newList(n_1_00),
		height2,
		newList(n_3_00, n_3_01, n_3_02),
	}

	n_2_00.Node().height = 1
	err := rh.sanityCheck()
	testutil.NotNil(t, err)
}

func Test_recomputeHeap_clear(t *testing.T) {
	g := New()
	rh := newRecomputeHeap(32)

	n50 := newHeightIncr(g, 5)
	n60 := newHeightIncr(g, 6)
	n70 := newHeightIncr(g, 7)

	rh.add(n50)
	rh.add(n60)
	rh.add(n70)

	testutil.Equal(t, 3, rh.numItems)
	testutil.Equal(t, 1, rh.heights[5].len())
	testutil.Equal(t, 1, rh.heights[6].len())
	testutil.Equal(t, 1, rh.heights[7].len())
	testutil.Equal(t, 5, rh.minHeight)
	testutil.Equal(t, 7, rh.maxHeight)

	rh.clear()

	testutil.Equal(t, 0, rh.numItems)
	testutil.Equal(t, 0, rh.heights[5].len())
	testutil.Equal(t, 0, rh.heights[6].len())
	testutil.Equal(t, 0, rh.heights[7].len())
	testutil.Equal(t, 0, rh.minHeight)
	testutil.Equal(t, 0, rh.maxHeight)

	rh.add(n50)
	rh.add(n60)
	rh.add(n70)

	testutil.Equal(t, 3, rh.numItems)
	testutil.Equal(t, 1, rh.heights[5].len())
	testutil.Equal(t, 1, rh.heights[6].len())
	testutil.Equal(t, 1, rh.heights[7].len())
	testutil.Equal(t, 7, rh.maxHeight)
}

func Test_recomputeHeap_removeMinUnsafe(t *testing.T) {
	g := New()

	rh := newRecomputeHeap(10)

	n00 := newHeightIncr(g, 0)
	n01 := newHeightIncr(g, 0)
	n02 := newHeightIncr(g, 0)

	n10 := newHeightIncr(g, 1)
	n11 := newHeightIncr(g, 1)
	n12 := newHeightIncr(g, 1)
	n13 := newHeightIncr(g, 1)

	n50 := newHeightIncr(g, 5)
	n51 := newHeightIncr(g, 5)
	n52 := newHeightIncr(g, 5)
	n53 := newHeightIncr(g, 5)
	n54 := newHeightIncr(g, 5)

	rh.add(n00)
	rh.add(n01)
	rh.add(n02)
	rh.add(n10)
	rh.add(n11)
	rh.add(n12)
	rh.add(n13)
	rh.add(n50)
	rh.add(n51)
	rh.add(n52)
	rh.add(n53)
	rh.add(n54)

	node, ok := rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n00.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n01.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n02.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n10.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n11.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n12.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n13.Node().id, node.Node().id)

	rh.add(n10)
	rh.add(n11)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n10.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n11.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n50.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n51.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n52.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n53.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, n54.Node().id, node.Node().id)

	node, ok = rh.removeMinUnsafe()
	testutil.Equal(t, false, ok)
	testutil.Nil(t, node)
}
