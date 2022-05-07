package incr

import (
	"testing"
)

func Test_recomputeHeap(t *testing.T) {
	rh := newRecomputeHeap(32)

	n0 := newHeightIncr(0)
	n1 := newHeightIncr(0)
	rh.add(n0)
	ItsEqual(t, 1, rh.len)
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 1, len(rh.heights[0]))
	rh.add(n1)
	ItsEqual(t, 2, rh.len)
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 2, len(rh.heights[0]))

	n2 := newHeightIncr(1)
	rh.add(n2)
	ItsEqual(t, 3, rh.len)
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 2, len(rh.heights[0]))
	ItsEqual(t, 1, len(rh.heights[1]))

	n3 := newHeightIncr(10)
	rh.add(n3)
	ItsEqual(t, 4, rh.len)
	ItsEqual(t, 32, len(rh.heights))
	ItsEqual(t, 2, len(rh.heights[0]))
	ItsEqual(t, 1, len(rh.heights[1]))
	ItsEqual(t, 1, len(rh.heights[10]))

	for rh.len > 0 {
		n := rh.removeMin()
		ItsNotNil(t, n)
		ItsNotNil(t, n.Node())
	}
	ItsEqual(t, 0, rh.len)

	rh.add(n0)
	rh.add(n1)
	rh.add(n2)
	rh.add(n3)

	rh.remove(n1)

	ItsEqual(t, true, rh.has(n0))
	ItsEqual(t, false, rh.has(n1))
	ItsEqual(t, true, rh.has(n2))
	ItsEqual(t, true, rh.has(n3))
}

func newHeightIncr(height int) *heightIncr {
	return &heightIncr{
		n: &Node{
			id:     newIdentifier(),
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
