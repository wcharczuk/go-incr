package incr

import (
	"testing"
)

func Test_recomputeHeap(t *testing.T) {
	rh := newRecomputeHeap(32)

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	rh.Add(n0)
	ItsEqual(t, 1, rh.Len())
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 1, rh.heights[0].len)
	rh.Add(n1)
	ItsEqual(t, 2, rh.Len())
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 2, rh.heights[0].len)

	n2 := newHeightIncr(1)
	rh.Add(n2)
	ItsEqual(t, 3, rh.Len())
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 2, rh.heights[0].len)
	ItsEqual(t, 1, rh.heights[1].len)

	n3 := newHeightIncr(10)
	rh.Add(n3)
	ItsEqual(t, 4, rh.Len())
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 2, rh.heights[0].len)
	ItsEqual(t, 1, rh.heights[1].len)
	ItsEqual(t, 1, rh.heights[10].len)

	for rh.Len() > 0 {
		n := rh.RemoveMin()
		ItsNotNil(t, n)
		ItsNotNil(t, n.Node())
	}
	ItsEqual(t, 0, rh.Len())

	rh.Add(n0)
	rh.Add(n1)
	rh.Add(n2)
	rh.Add(n3)

	rh.Remove(n1)

	ItsEqual(t, true, rh.Has(n0))
	ItsEqual(t, false, rh.Has(n1))
	ItsEqual(t, true, rh.Has(n2))
	ItsEqual(t, true, rh.Has(n3))
}

func newHeightIncr(height int) *heightIncr {
	return &heightIncr{
		n: &Node{
			id:     NewIdentifier(),
			height: height,
		},
	}
}

type heightIncr struct {
	Incr[struct{}]
	n *Node
}

func (hi heightIncr) Node() *Node {
	return hi.n
}

func Test_recomputeHeap_removeMinHeight(t *testing.T) {
	n00 := newHeightIncr(0)
	n01 := newHeightIncr(0)
	n02 := newHeightIncr(0)

	n10 := newHeightIncr(1)
	n11 := newHeightIncr(1)
	n12 := newHeightIncr(1)
	n13 := newHeightIncr(1)

	rh := newRecomputeHeap(2)
	rh.Add(n00)
	rh.Add(n01)
	rh.Add(n02)

	rh.Add(n10)
	rh.Add(n11)
	rh.Add(n12)
	rh.Add(n13)

	output := rh.RemoveMinHeight()
	ItsEqual(t, 3, len(output))
	ItsNil(t, rh.heights[0].head)
	ItsNil(t, rh.heights[0].tail)
	ItsEqual(t, 0, rh.heights[0].len)
}

func Test_recomputeHeapList_popAll(t *testing.T) {
	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	n2 := newHeightIncr(0)

	rhl := new(recomputeHeapList)
	rhl.push(n0)
	rhl.push(n1)
	rhl.push(n2)

	ItsNotNil(t, rhl.head)
	ItsNotNil(t, rhl.tail)
	ItsEqual(t, 3, rhl.len)

	output := rhl.popAll()
	ItsEqual(t, 3, len(output))
	ItsEqual(t, 0, rhl.len)
	ItsNil(t, rhl.head)
	ItsNil(t, rhl.tail)
}
