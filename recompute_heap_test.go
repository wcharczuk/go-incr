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
