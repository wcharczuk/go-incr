package incr

import (
	"testing"

	. "github.com/wcharczuk/go-incr/testutil"
)

func Test_recomputeHeap_Add(t *testing.T) {

	rh := newRecomputeHeap(32)

	n50 := newHeightIncr(5)
	n60 := newHeightIncr(6)
	n70 := newHeightIncr(7)

	rh.Add(n50)

	// assertions post add n50
	{
		ItsEqual(t, 1, rh.Len())
		ItsEqual(t, 1, len(rh.heights[5]))
		ItsEqual(t, true, rh.Has(n50))
		ItsEqual(t, false, rh.Has(n60))
		ItsEqual(t, false, rh.Has(n70))
		ItsEqual(t, 5, rh.MinHeight())
		ItsEqual(t, 5, rh.MaxHeight())
	}

	rh.Add(n60)

	// assertions post add n60
	{
		ItsEqual(t, 2, rh.Len())
		ItsEqual(t, 1, len(rh.heights[5]))
		ItsEqual(t, 1, len(rh.heights[6]))
		ItsEqual(t, true, rh.Has(n50))
		ItsEqual(t, true, rh.Has(n50))
		ItsEqual(t, false, rh.Has(n70))
		ItsEqual(t, 5, rh.MinHeight())
		ItsEqual(t, 6, rh.MaxHeight())
	}

	rh.Add(n70)

	// assertions post add n70
	{
		ItsEqual(t, 3, rh.Len())
		ItsEqual(t, 1, len(rh.heights[5]))
		ItsEqual(t, 1, len(rh.heights[6]))
		ItsEqual(t, 1, len(rh.heights[7]))
		ItsEqual(t, true, rh.Has(n50))
		ItsEqual(t, true, rh.Has(n60))
		ItsEqual(t, true, rh.Has(n70))
		ItsEqual(t, 5, rh.MinHeight())
		ItsEqual(t, 7, rh.MaxHeight())
	}
}

func Test_recomputeHeap_RemoveMinHeight(t *testing.T) {
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

	rh.Add(n00)
	rh.Add(n01)
	rh.Add(n02)
	rh.Add(n10)
	rh.Add(n11)
	rh.Add(n12)
	rh.Add(n13)
	rh.Add(n50)
	rh.Add(n51)
	rh.Add(n52)
	rh.Add(n53)
	rh.Add(n54)

	ItsEqual(t, 12, rh.Len())
	ItsEqual(t, 0, rh.MinHeight())
	ItsEqual(t, 5, rh.MaxHeight())

	output := rh.RemoveMinHeight()
	ItsNil(t, rh.sanityCheck())
	ItsEqual(t, 9, rh.Len())
	ItsEqual(t, 3, len(output))
	ItsEqual(t, 0, len(rh.heights[0]))
	ItsEqual(t, 4, len(rh.heights[1]))
	ItsEqual(t, 5, len(rh.heights[5]))

	ItsEqual(t, 1, rh.MinHeight())
	ItsEqual(t, 5, rh.MaxHeight())

	output = rh.RemoveMinHeight()
	ItsNil(t, rh.sanityCheck())
	ItsEqual(t, 5, rh.Len())
	ItsEqual(t, 4, len(output))
	ItsEqual(t, 0, len(rh.heights[0]))
	ItsEqual(t, 0, len(rh.heights[1]))
	ItsEqual(t, 5, len(rh.heights[5]))

	output = rh.RemoveMinHeight()
	ItsNil(t, rh.sanityCheck())
	ItsEqual(t, 0, rh.Len())
	ItsEqual(t, 5, len(output))
	ItsEqual(t, 0, len(rh.heights[0]))
	ItsEqual(t, 0, len(rh.heights[1]))
	ItsEqual(t, 0, len(rh.heights[5]))

	rh.Add(n50)
	rh.Add(n51)
	rh.Add(n52)
	rh.Add(n53)
	rh.Add(n54)

	ItsNil(t, rh.sanityCheck())
	ItsEqual(t, 5, rh.Len())
	ItsEqual(t, 5, rh.MinHeight())
	ItsEqual(t, 5, rh.MaxHeight())

	output = rh.RemoveMinHeight()
	ItsNil(t, rh.sanityCheck())
	ItsEqual(t, 0, rh.Len())
	ItsEqual(t, 5, len(output))
	ItsEqual(t, 0, len(rh.heights[0]))
	ItsEqual(t, 0, len(rh.heights[1]))
	ItsEqual(t, 0, len(rh.heights[5]))
}

func Test_recomputeHeap_Remove(t *testing.T) {
	rh := newRecomputeHeap(10)
	n10 := newHeightIncr(1)
	n11 := newHeightIncr(1)
	n20 := newHeightIncr(2)
	n21 := newHeightIncr(2)
	n22 := newHeightIncr(2)
	n30 := newHeightIncr(3)

	// this should just return
	rh.Remove(n10)

	rh.Add(n10)
	rh.Add(n11)
	rh.Add(n20)
	rh.Add(n21)
	rh.Add(n22)
	rh.Add(n30)

	ItsNil(t, rh.sanityCheck())
	ItsEqual(t, true, rh.Has(n10))
	ItsEqual(t, true, rh.Has(n11))
	ItsEqual(t, true, rh.Has(n20))
	ItsEqual(t, true, rh.Has(n21))
	ItsEqual(t, true, rh.Has(n22))
	ItsEqual(t, true, rh.Has(n30))

	rh.Remove(n21)

	ItsNil(t, rh.sanityCheck())
	ItsEqual(t, 5, rh.Len())
	ItsEqual(t, true, rh.Has(n10))
	ItsEqual(t, true, rh.Has(n11))
	ItsEqual(t, true, rh.Has(n20))
	ItsEqual(t, false, rh.Has(n21))
	ItsEqual(t, true, rh.Has(n22))
	ItsEqual(t, true, rh.Has(n30))

	ItsEqual(t, 2, len(rh.heights[2]))

	rh.Remove(n10)
	rh.Remove(n11)

	ItsNil(t, rh.sanityCheck())
	ItsEqual(t, 3, rh.Len())
	ItsEqual(t, 0, len(rh.heights[1]))
	ItsEqual(t, 2, rh.MinHeight())
	ItsEqual(t, 3, rh.MaxHeight())
}

func Test_recomputeHeap_nextMinHeightUnsafe_noItems(t *testing.T) {
	rh := new(recomputeHeap)

	rh.minHeight = 1
	rh.maxHeight = 3

	next := rh.nextMinHeightUnsafe()
	ItsEqual(t, 0, next)
}

func Test_recomputeHeap_nextMinHeightUnsafe_pastMax(t *testing.T) {
	r0 := Return(testContext(), "hello")
	rh := newRecomputeHeap(4)
	rh.minHeight = 1
	rh.maxHeight = 3

	rh.lookup[r0.Node().id] = &recomputeHeapItem{node: r0, height: 0}
	next := rh.nextMinHeightUnsafe()
	ItsEqual(t, 0, next)
}

func Test_recomputeHeap_maybeAddNewHeights(t *testing.T) {
	rh := newRecomputeHeap(8)
	ItsEqual(t, 8, len(rh.heights))
	rh.maybeAddNewHeights(9) // we use (1) indexing!
	ItsEqual(t, 10, len(rh.heights))
}

func Test_recomputeHeap_Add_adjustsHeights(t *testing.T) {
	rh := newRecomputeHeap(8)
	ItsEqual(t, 8, len(rh.heights))

	v0 := newHeightIncr(32)
	rh.Add(v0)
	ItsEqual(t, 33, len(rh.heights))
	ItsEqual(t, 32, rh.minHeight)
	ItsEqual(t, 32, rh.maxHeight)

	v1 := newHeightIncr(64)
	rh.Add(v1)
	ItsEqual(t, 65, len(rh.heights))
	ItsEqual(t, 32, rh.minHeight)
	ItsEqual(t, 64, rh.maxHeight)
}

func Test_recomputeHeap_Add_regression2(t *testing.T) {
	// another real world use case! also insane!
	rh := newRecomputeHeap(256)

	observer4945d288 := newHeightIncr(1)
	rh.Add(observer4945d288)
	observer87df48be := newHeightIncr(1)
	rh.Add(observer87df48be)
	mapf2cb6e46 := newHeightIncr(0)
	rh.Add(mapf2cb6e46)
	map26e9bfb2a := newHeightIncr(0)
	rh.Add(map26e9bfb2a)
	map26e9bfb2a.n.height = 2
	rh.Add(map26e9bfb2a)
	map2dfe7c676 := newHeightIncr(1)
	rh.Add(map2dfe7c676)
	map2aa9d55f9 := newHeightIncr(1)
	rh.Add(map2aa9d55f9)
	observerbaad6dd3 := newHeightIncr(1)
	rh.Add(observerbaad6dd3)
	map2aa3f9a14 := newHeightIncr(1)
	rh.Add(map2aa3f9a14)
	observer6e9e8864 := newHeightIncr(1)
	rh.Add(observer6e9e8864)
	varb35bfa8a := newHeightIncr(1)
	rh.Add(varb35bfa8a)
	var54b93408 := newHeightIncr(1)
	rh.Add(var54b93408)
	alwaysc83986c6 := newHeightIncr(0)
	rh.Add(alwaysc83986c6)
	cutoff9d454a57 := newHeightIncr(0)
	rh.Add(cutoff9d454a57)
	varc0898518 := newHeightIncr(1)
	rh.Add(varc0898518)
	cutoff9d454a57.n.height = 4
	rh.Add(cutoff9d454a57)
	mapf2cb6e46.n.height = 3
	rh.Add(mapf2cb6e46)
	alwaysc83986c6.n.height = 2
	rh.Add(alwaysc83986c6)
	map26e9bfb2a.n.height = 7
	rh.Add(map26e9bfb2a)
	map2aa3f9a14.n.height = 5
	rh.Add(map2aa3f9a14)
	observer6e9e8864.n.height = 6
	rh.Add(observer6e9e8864)
	map2dfe7c676.n.height = 6
	rh.Add(map2dfe7c676)
	map2aa9d55f9.n.height = 6
	rh.Add(map2aa9d55f9)
	observerbaad6dd3.n.height = 8
	rh.Add(observerbaad6dd3)

	// now start """stabilization"""

	ItsEqual(t, 1, rh.minHeight)
	ItsEqual(t, 5, len(rh.heights[1]))

	minHeightBlock := rh.RemoveMinHeight()
	ItsEqual(t, 5, len(minHeightBlock))
	ItsEqual(t, true, allHeight(minHeightBlock, 1))
}

func Test_recomputeHeap_Fix(t *testing.T) {
	rh := newRecomputeHeap(8)
	v0 := newHeightIncr(2)
	rh.Add(v0)
	v1 := newHeightIncr(3)
	rh.Add(v1)
	v2 := newHeightIncr(4)
	rh.Add(v2)

	ItsEqual(t, 2, rh.minHeight)
	ItsEqual(t, 1, len(rh.heights[2]))
	ItsEqual(t, 1, len(rh.heights[3]))
	ItsEqual(t, 1, len(rh.heights[4]))
	ItsEqual(t, 4, rh.maxHeight)

	v0.n.height = 1
	rh.Fix(v0.n.id)

	ItsEqual(t, 1, rh.minHeight)
	ItsEqual(t, 1, len(rh.heights[1]))
	ItsEqual(t, 0, len(rh.heights[2]))
	ItsEqual(t, 1, len(rh.heights[3]))
	ItsEqual(t, 1, len(rh.heights[4]))
	ItsEqual(t, 4, rh.maxHeight)

	rh.Fix(v0.n.id)
	ItsEqual(t, 1, rh.minHeight)
	ItsEqual(t, 1, len(rh.heights[1]))
	ItsEqual(t, 0, len(rh.heights[2]))
	ItsEqual(t, 1, len(rh.heights[3]))
	ItsEqual(t, 1, len(rh.heights[4]))
	ItsEqual(t, 4, rh.maxHeight)

	v2.n.height = 5
	rh.Fix(v2.n.id)

	ItsEqual(t, 1, rh.minHeight)
	ItsEqual(t, 1, len(rh.heights[1]))
	ItsEqual(t, 0, len(rh.heights[2]))
	ItsEqual(t, 1, len(rh.heights[3]))
	ItsEqual(t, 0, len(rh.heights[4]))
	ItsEqual(t, 1, len(rh.heights[5]))
	ItsEqual(t, 5, rh.maxHeight)
}

func Test_recomputeHeap_sanityCheck_ok_badNodeHeight(t *testing.T) {
	rh := newRecomputeHeap(8)

	n_1_00 := newMockBareNodeWithHeight(1)
	n_2_00 := newMockBareNodeWithHeight(2)
	n_2_01 := newMockBareNodeWithHeight(2)
	n_3_00 := newMockBareNodeWithHeight(3)
	n_3_01 := newMockBareNodeWithHeight(3)
	n_3_02 := newMockBareNodeWithHeight(3)

	rh.heights = []map[Identifier]*recomputeHeapItem{
		nil,
		newList(n_1_00),
		newList(n_2_00, n_2_01),
		newList(n_3_00, n_3_01, n_3_02),
	}

	err := rh.sanityCheck()
	ItsNil(t, err)

	n_3_00.n.height = 2

	err = rh.sanityCheck()
	ItsNotNil(t, err)
}

func Test_recomputeHeap_sanityCheck_badItemHeight(t *testing.T) {
	rh := newRecomputeHeap(8)

	n_1_00 := newMockBareNodeWithHeight(1)
	n_2_00 := newMockBareNodeWithHeight(2)
	n_2_01 := newMockBareNodeWithHeight(2)
	n_3_00 := newMockBareNodeWithHeight(3)
	n_3_01 := newMockBareNodeWithHeight(3)
	n_3_02 := newMockBareNodeWithHeight(3)

	height2, height2Items := newListWithItems(n_2_00, n_2_01)

	rh.heights = []map[Identifier]*recomputeHeapItem{
		nil,
		newList(n_1_00),
		height2,
		newList(n_3_00, n_3_01, n_3_02),
	}

	height2Items[0].height = 1
	err := rh.sanityCheck()
	ItsNotNil(t, err)
}
