package incr

import (
	"testing"

	. "github.com/wcharczuk/go-incr/testutil"
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
		Equal(t, 1, rh.len())
		Equal(t, 1, rh.heights[5].len())
		Equal(t, true, rh.has(n50))
		Equal(t, false, rh.has(n60))
		Equal(t, false, rh.has(n70))
		Equal(t, 5, rh.minHeight)
		Equal(t, 5, rh.maxHeight)
	}

	rh.add(n60)

	// assertions post add n60
	{
		Equal(t, 2, rh.len())
		Equal(t, 1, rh.heights[5].len())
		Equal(t, 1, rh.heights[6].len())
		Equal(t, true, rh.has(n50))
		Equal(t, true, rh.has(n50))
		Equal(t, false, rh.has(n70))
		Equal(t, 5, rh.minHeight)
		Equal(t, 6, rh.maxHeight)
	}

	rh.add(n70)

	// assertions post add n70
	{
		Equal(t, 3, rh.len())
		Equal(t, 1, rh.heights[5].len())
		Equal(t, 1, rh.heights[6].len())
		Equal(t, 1, rh.heights[7].len())
		Equal(t, true, rh.has(n50))
		Equal(t, true, rh.has(n60))
		Equal(t, true, rh.has(n70))
		Equal(t, 5, rh.minHeight)
		Equal(t, 7, rh.maxHeight)
	}
}

func Test_recomputeHeap_removeMinHeight(t *testing.T) {
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

	Equal(t, 12, rh.len())
	Equal(t, 0, rh.minHeight)
	Equal(t, 5, rh.maxHeight)

	output := rh.removeMinHeight()
	Nil(t, rh.sanityCheck())
	Equal(t, 9, rh.len())
	Equal(t, 3, len(output))
	Equal(t, 0, rh.heights[0].len())
	Equal(t, 4, rh.heights[1].len())
	Equal(t, 5, rh.heights[5].len())

	Equal(t, 1, rh.minHeight)
	Equal(t, 5, rh.maxHeight)

	output = rh.removeMinHeight()
	Nil(t, rh.sanityCheck())
	Equal(t, 5, rh.len())
	Equal(t, 4, len(output))
	Equal(t, 0, rh.heights[0].len())
	Equal(t, 0, rh.heights[1].len())
	Equal(t, 5, rh.heights[5].len())

	output = rh.removeMinHeight()
	Nil(t, rh.sanityCheck())
	Equal(t, 0, rh.len())
	Equal(t, 5, len(output))
	Equal(t, 0, rh.heights[0].len())
	Equal(t, 0, rh.heights[1].len())
	Equal(t, 0, rh.heights[5].len())

	rh.add(n50)
	rh.add(n51)
	rh.add(n52)
	rh.add(n53)
	rh.add(n54)

	Nil(t, rh.sanityCheck())
	Equal(t, 5, rh.len())
	Equal(t, 5, rh.minHeight)
	Equal(t, 5, rh.maxHeight)

	output = rh.removeMinHeight()
	Nil(t, rh.sanityCheck())
	Equal(t, 0, rh.len())
	Equal(t, 5, len(output))
	Equal(t, 0, rh.heights[0].len())
	Equal(t, 0, rh.heights[1].len())
	Equal(t, 0, rh.heights[5].len())
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

	// this should just return
	rh.remove(n10)

	rh.add(n10)
	rh.add(n11)
	rh.add(n20)
	rh.add(n21)
	rh.add(n22)
	rh.add(n30)

	Nil(t, rh.sanityCheck())
	Equal(t, true, rh.has(n10))
	Equal(t, true, rh.has(n11))
	Equal(t, true, rh.has(n20))
	Equal(t, true, rh.has(n21))
	Equal(t, true, rh.has(n22))
	Equal(t, true, rh.has(n30))

	rh.remove(n21)

	Nil(t, rh.sanityCheck())
	Equal(t, 5, rh.len())
	Equal(t, true, rh.has(n10))
	Equal(t, true, rh.has(n11))
	Equal(t, true, rh.has(n20))
	Equal(t, false, rh.has(n21))

	for _, h := range rh.heights {
		if h == nil {
			continue
		}
		Equal(t, false, h.has(n21.n.id))
	}

	Equal(t, true, rh.has(n22))
	Equal(t, true, rh.has(n30))

	Equal(t, 2, rh.heights[2].len())

	rh.remove(n10)
	rh.remove(n11)

	Nil(t, rh.sanityCheck())
	Equal(t, 3, rh.len())
	Equal(t, 0, rh.heights[1].len())
	Equal(t, 2, rh.minHeight)
	Equal(t, 3, rh.maxHeight)

	for _, h := range rh.heights {
		if h == nil {
			continue
		}
		Equal(t, false, h.has(n10.n.id))
	}
	for _, h := range rh.heights {
		if h == nil {
			continue
		}
		Equal(t, false, h.has(n11.n.id))
	}
}

func Test_recomputeHeap_nextMinHeightUnsafe_noItems(t *testing.T) {
	rh := new(recomputeHeap)

	rh.minHeight = 1
	rh.maxHeight = 3

	next := rh.nextMinHeightUnsafe()
	Equal(t, 0, next)
}

func Test_recomputeHeap_nextMinHeightUnsafe_pastMax(t *testing.T) {
	g := New()
	r0 := Return(g, "hello")
	rh := newRecomputeHeap(4)
	rh.minHeight = 1
	rh.maxHeight = 3

	rh.lookup[r0.Node().id] = r0
	next := rh.nextMinHeightUnsafe()
	Equal(t, 0, next)
}

func Test_recomputeHeap_maybeAddNewHeights(t *testing.T) {
	rh := newRecomputeHeap(8)
	Equal(t, 8, len(rh.heights))
	rh.maybeAddNewHeights(9) // we use (1) indexing!
	Equal(t, 10, len(rh.heights))
}

func Test_recomputeHeap_add_adjustsHeights(t *testing.T) {
	g := New()
	rh := newRecomputeHeap(8)
	Equal(t, 8, len(rh.heights))

	v0 := newHeightIncr(g, 32)
	rh.add(v0)
	Equal(t, 33, len(rh.heights))
	Equal(t, 32, rh.minHeight)
	Equal(t, 32, rh.maxHeight)

	v1 := newHeightIncr(g, 64)
	rh.add(v1)
	Equal(t, 65, len(rh.heights))
	Equal(t, 32, rh.minHeight)
	Equal(t, 64, rh.maxHeight)
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

	Equal(t, 2, rh.minHeight)
	Equal(t, 1, rh.heights[2].len())
	Equal(t, 1, rh.heights[3].len())
	Equal(t, 1, rh.heights[4].len())
	Equal(t, 4, rh.maxHeight)

	v0.n.height = 1
	rh.fix(v0.n.id)

	Equal(t, 1, rh.minHeight)
	Equal(t, 1, rh.heights[1].len())
	Equal(t, 0, rh.heights[2].len())
	Equal(t, 1, rh.heights[3].len())
	Equal(t, 1, rh.heights[4].len())
	Equal(t, 4, rh.maxHeight)

	rh.fix(v0.n.id)
	Equal(t, 1, rh.minHeight)
	Equal(t, 1, rh.heights[1].len())
	Equal(t, 0, rh.heights[2].len())
	Equal(t, 1, rh.heights[3].len())
	Equal(t, 1, rh.heights[4].len())
	Equal(t, 4, rh.maxHeight)

	v2.n.height = 5
	rh.fix(v2.n.id)

	Equal(t, 1, rh.minHeight)
	Equal(t, 1, rh.heights[1].len())
	Equal(t, 0, rh.heights[2].len())
	Equal(t, 1, rh.heights[3].len())
	Equal(t, 0, rh.heights[4].len())
	Equal(t, 1, rh.heights[5].len())
	Equal(t, 5, rh.maxHeight)
}

func Test_recomputeHeap_sanityCheck_ok_badNodeHeight(t *testing.T) {
	rh := newRecomputeHeap(8)

	n_1_00 := newMockBareNodeWithHeight(1)
	n_2_00 := newMockBareNodeWithHeight(2)
	n_2_01 := newMockBareNodeWithHeight(2)
	n_3_00 := newMockBareNodeWithHeight(3)
	n_3_01 := newMockBareNodeWithHeight(3)
	n_3_02 := newMockBareNodeWithHeight(3)

	rh.heights = []*list[Identifier, INode]{
		nil,
		newList(n_1_00),
		newList(n_2_00, n_2_01),
		newList(n_3_00, n_3_01, n_3_02),
	}
	rh.minHeight = 1
	rh.lookup[n_1_00.n.id] = n_1_00
	rh.lookup[n_2_00.n.id] = n_2_00
	rh.lookup[n_2_01.n.id] = n_2_01
	rh.lookup[n_3_00.n.id] = n_3_00
	rh.lookup[n_3_01.n.id] = n_3_01
	rh.lookup[n_3_02.n.id] = n_3_02

	err := rh.sanityCheck()
	Nil(t, err)

	n_3_00.n.height = 2

	err = rh.sanityCheck()
	NotNil(t, err)
}

func Test_recomputeHeap_sanityCheck_badItemHeight(t *testing.T) {
	rh := newRecomputeHeap(8)

	n_1_00 := newMockBareNodeWithHeight(1)
	n_2_00 := newMockBareNodeWithHeight(2)
	n_2_01 := newMockBareNodeWithHeight(2)
	n_3_00 := newMockBareNodeWithHeight(3)
	n_3_01 := newMockBareNodeWithHeight(3)
	n_3_02 := newMockBareNodeWithHeight(3)

	height2 := newList(n_2_00, n_2_01)

	rh.heights = []*list[Identifier, INode]{
		nil,
		newList(n_1_00),
		height2,
		newList(n_3_00, n_3_01, n_3_02),
	}
	rh.lookup[n_1_00.n.id] = n_1_00
	rh.lookup[n_2_00.n.id] = n_2_00
	rh.lookup[n_2_01.n.id] = n_2_01
	rh.lookup[n_3_00.n.id] = n_3_00
	rh.lookup[n_3_01.n.id] = n_3_01
	rh.lookup[n_3_02.n.id] = n_3_02

	n_2_00.Node().heightInRecomputeHeap = 1
	err := rh.sanityCheck()
	NotNil(t, err)
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

	Equal(t, 3, len(rh.lookup))
	Equal(t, 1, rh.heights[5].len())
	Equal(t, 1, rh.heights[6].len())
	Equal(t, 1, rh.heights[7].len())
	Equal(t, 5, rh.minHeight)
	Equal(t, 7, rh.maxHeight)

	rh.clear()

	Equal(t, 0, len(rh.lookup))
	Equal(t, 0, rh.heights[5].len())
	Equal(t, 0, rh.heights[6].len())
	Equal(t, 0, rh.heights[7].len())
	Equal(t, 0, rh.minHeight)
	Equal(t, 0, rh.maxHeight)

	rh.add(n50)
	rh.add(n60)
	rh.add(n70)

	Equal(t, 3, len(rh.lookup))
	Equal(t, 1, rh.heights[5].len())
	Equal(t, 1, rh.heights[6].len())
	Equal(t, 1, rh.heights[7].len())
	Equal(t, 7, rh.maxHeight)
}
