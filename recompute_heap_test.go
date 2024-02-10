package incr

import (
	"testing"

	. "github.com/wcharczuk/go-incr/testutil"
)

func Test_recomputeHeap_add(t *testing.T) {
	rh := newRecomputeHeap(32)

	n50 := newHeightIncr(5)
	n60 := newHeightIncr(6)
	n70 := newHeightIncr(7)

	rh.add(n50)

	// assertions post add n50
	{
		Equal(t, 1, rh.len())
		Equal(t, 1, len(rh.heights[5]))
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
		Equal(t, 1, len(rh.heights[5]))
		Equal(t, 1, len(rh.heights[6]))
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
		Equal(t, 1, len(rh.heights[5]))
		Equal(t, 1, len(rh.heights[6]))
		Equal(t, 1, len(rh.heights[7]))
		Equal(t, true, rh.has(n50))
		Equal(t, true, rh.has(n60))
		Equal(t, true, rh.has(n70))
		Equal(t, 5, rh.minHeight)
		Equal(t, 7, rh.maxHeight)
	}
}

func Test_recomputeHeap_removeMinHeight(t *testing.T) {
	rh := newRecomputeHeap(10)

	n00 := newHeightIncr(0)
	n01 := newHeightIncr(0)
	n02 := newHeightIncr(0)

	n10 := newHeightIncr(1)
	n11 := newHeightIncr(1)
	n12 := newHeightIncr(1)
	n13 := newHeightIncr(1)

	n50 := newHeightIncr(5)
	n51 := newHeightIncr(5)
	n52 := newHeightIncr(5)
	n53 := newHeightIncr(5)
	n54 := newHeightIncr(5)

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
	Equal(t, 0, len(rh.heights[0]))
	Equal(t, 4, len(rh.heights[1]))
	Equal(t, 5, len(rh.heights[5]))

	Equal(t, 1, rh.minHeight)
	Equal(t, 5, rh.maxHeight)

	output = rh.removeMinHeight()
	Nil(t, rh.sanityCheck())
	Equal(t, 5, rh.len())
	Equal(t, 4, len(output))
	Equal(t, 0, len(rh.heights[0]))
	Equal(t, 0, len(rh.heights[1]))
	Equal(t, 5, len(rh.heights[5]))

	output = rh.removeMinHeight()
	Nil(t, rh.sanityCheck())
	Equal(t, 0, rh.len())
	Equal(t, 5, len(output))
	Equal(t, 0, len(rh.heights[0]))
	Equal(t, 0, len(rh.heights[1]))
	Equal(t, 0, len(rh.heights[5]))

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
	Equal(t, 0, len(rh.heights[0]))
	Equal(t, 0, len(rh.heights[1]))
	Equal(t, 0, len(rh.heights[5]))
}

func Test_recomputeHeap_remove(t *testing.T) {
	rh := newRecomputeHeap(10)
	n10 := newHeightIncr(1)
	n11 := newHeightIncr(1)
	n20 := newHeightIncr(2)
	n21 := newHeightIncr(2)
	n22 := newHeightIncr(2)
	n30 := newHeightIncr(3)

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
		_, ok := h[n21.n.id]
		Equal(t, false, ok)
	}

	Equal(t, true, rh.has(n22))
	Equal(t, true, rh.has(n30))

	Equal(t, 2, len(rh.heights[2]))

	rh.remove(n10)
	rh.remove(n11)

	Nil(t, rh.sanityCheck())
	Equal(t, 3, rh.len())
	Equal(t, 0, len(rh.heights[1]))
	Equal(t, 2, rh.minHeight)
	Equal(t, 3, rh.maxHeight)

	for _, h := range rh.heights {
		_, ok := h[n10.n.id]
		Equal(t, false, ok)
	}
	for _, h := range rh.heights {
		_, ok := h[n11.n.id]
		Equal(t, false, ok)
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
	r0 := Return(Root(), "hello")
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
	rh := newRecomputeHeap(8)
	Equal(t, 8, len(rh.heights))

	v0 := newHeightIncr(32)
	rh.add(v0)
	Equal(t, 33, len(rh.heights))
	Equal(t, 32, rh.minHeight)
	Equal(t, 32, rh.maxHeight)

	v1 := newHeightIncr(64)
	rh.add(v1)
	Equal(t, 65, len(rh.heights))
	Equal(t, 32, rh.minHeight)
	Equal(t, 64, rh.maxHeight)
}

func Test_recomputeHeap_add_regression2(t *testing.T) {
	// another real world use case! also insane!
	rh := newRecomputeHeap(256)

	observer4945d288 := newHeightIncr(1)
	rh.add(observer4945d288)
	observer87df48be := newHeightIncr(1)
	rh.add(observer87df48be)
	mapf2cb6e46 := newHeightIncr(0)
	rh.add(mapf2cb6e46)
	map26e9bfb2a := newHeightIncr(0)
	rh.add(map26e9bfb2a)
	map26e9bfb2a.n.height = 2
	rh.add(map26e9bfb2a)
	map2dfe7c676 := newHeightIncr(1)
	rh.add(map2dfe7c676)
	map2aa9d55f9 := newHeightIncr(1)
	rh.add(map2aa9d55f9)
	observerbaad6dd3 := newHeightIncr(1)
	rh.add(observerbaad6dd3)
	map2aa3f9a14 := newHeightIncr(1)
	rh.add(map2aa3f9a14)
	observer6e9e8864 := newHeightIncr(1)
	rh.add(observer6e9e8864)
	varb35bfa8a := newHeightIncr(1)
	rh.add(varb35bfa8a)
	var54b93408 := newHeightIncr(1)
	rh.add(var54b93408)
	alwaysc83986c6 := newHeightIncr(0)
	rh.add(alwaysc83986c6)
	cutoff9d454a57 := newHeightIncr(0)
	rh.add(cutoff9d454a57)
	varc0898518 := newHeightIncr(1)
	rh.add(varc0898518)
	cutoff9d454a57.n.height = 4
	rh.add(cutoff9d454a57)
	mapf2cb6e46.n.height = 3
	rh.add(mapf2cb6e46)
	alwaysc83986c6.n.height = 2
	rh.add(alwaysc83986c6)
	map26e9bfb2a.n.height = 7
	rh.add(map26e9bfb2a)
	map2aa3f9a14.n.height = 5
	rh.add(map2aa3f9a14)
	observer6e9e8864.n.height = 6
	rh.add(observer6e9e8864)
	map2dfe7c676.n.height = 6
	rh.add(map2dfe7c676)
	map2aa9d55f9.n.height = 6
	rh.add(map2aa9d55f9)
	observerbaad6dd3.n.height = 8
	rh.add(observerbaad6dd3)

	// now start """stabilization"""

	Equal(t, 1, rh.minHeight)
	Equal(t, 5, len(rh.heights[1]))

	minHeightBlock := rh.removeMinHeight()
	Equal(t, 5, len(minHeightBlock))
	Equal(t, true, allHeight(minHeightBlock, 1))
}

func Test_recomputeHeap_fix(t *testing.T) {
	rh := newRecomputeHeap(8)
	v0 := newHeightIncr(2)
	rh.add(v0)
	v1 := newHeightIncr(3)
	rh.add(v1)
	v2 := newHeightIncr(4)
	rh.add(v2)

	Equal(t, 2, rh.minHeight)
	Equal(t, 1, len(rh.heights[2]))
	Equal(t, 1, len(rh.heights[3]))
	Equal(t, 1, len(rh.heights[4]))
	Equal(t, 4, rh.maxHeight)

	v0.n.height = 1
	rh.fix(v0.n.id)

	Equal(t, 1, rh.minHeight)
	Equal(t, 1, len(rh.heights[1]))
	Equal(t, 0, len(rh.heights[2]))
	Equal(t, 1, len(rh.heights[3]))
	Equal(t, 1, len(rh.heights[4]))
	Equal(t, 4, rh.maxHeight)

	rh.fix(v0.n.id)
	Equal(t, 1, rh.minHeight)
	Equal(t, 1, len(rh.heights[1]))
	Equal(t, 0, len(rh.heights[2]))
	Equal(t, 1, len(rh.heights[3]))
	Equal(t, 1, len(rh.heights[4]))
	Equal(t, 4, rh.maxHeight)

	v2.n.height = 5
	rh.fix(v2.n.id)

	Equal(t, 1, rh.minHeight)
	Equal(t, 1, len(rh.heights[1]))
	Equal(t, 0, len(rh.heights[2]))
	Equal(t, 1, len(rh.heights[3]))
	Equal(t, 0, len(rh.heights[4]))
	Equal(t, 1, len(rh.heights[5]))
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

	rh.heights = []map[Identifier]INode{
		nil,
		newList(n_1_00),
		newList(n_2_00, n_2_01),
		newList(n_3_00, n_3_01, n_3_02),
	}

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

	rh.heights = []map[Identifier]INode{
		nil,
		newList(n_1_00),
		height2,
		newList(n_3_00, n_3_01, n_3_02),
	}

	n_2_00.Node().heightInRecomputeHeap = 1
	err := rh.sanityCheck()
	NotNil(t, err)
}

func Test_recomputeHeap_clear(t *testing.T) {
	rh := newRecomputeHeap(32)

	n50 := newHeightIncr(5)
	n60 := newHeightIncr(6)
	n70 := newHeightIncr(7)

	rh.add(n50)
	rh.add(n60)
	rh.add(n70)

	Equal(t, 3, len(rh.lookup))
	Equal(t, 1, len(rh.heights[5]))
	Equal(t, 1, len(rh.heights[6]))
	Equal(t, 1, len(rh.heights[7]))
	Equal(t, 5, rh.minHeight)
	Equal(t, 7, rh.maxHeight)

	rh.Clear()

	Equal(t, 0, len(rh.lookup))
	Equal(t, 0, len(rh.heights[5]))
	Equal(t, 0, len(rh.heights[6]))
	Equal(t, 0, len(rh.heights[7]))
	Equal(t, 0, rh.minHeight)
	Equal(t, 0, rh.maxHeight)

	rh.add(n50)
	rh.add(n60)
	rh.add(n70)

	Equal(t, 3, len(rh.lookup))
	Equal(t, 1, len(rh.heights[5]))
	Equal(t, 1, len(rh.heights[6]))
	Equal(t, 1, len(rh.heights[7]))
	Equal(t, 7, rh.maxHeight)
}
