package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Link_propagatesHeightChanges(t *testing.T) {
	n00 := newMockBareNode()
	n01 := newMockBareNode()
	n02 := newMockBareNode()

	n10 := newMockBareNode()
	n11 := newMockBareNode()

	Link(n00, n01, n02)
	Link(n10, n00)
	Link(n11, n10)

	testutil.ItsEqual(t, 2, n00.Node().height)
	testutil.ItsEqual(t, 1, n01.Node().height)
	testutil.ItsEqual(t, 1, n02.Node().height)

	testutil.ItsEqual(t, true, hasKey(n00.Node().parents, n01.Node().ID()))
	testutil.ItsEqual(t, true, hasKey(n00.Node().parents, n02.Node().ID()))

	testutil.ItsEqual(t, true, hasKey(n01.Node().children, n00.Node().ID()))
	testutil.ItsEqual(t, true, hasKey(n02.Node().children, n00.Node().ID()))

	testutil.ItsEqual(t, 3, n10.Node().height)
	testutil.ItsEqual(t, 4, n11.Node().height)
}

func Test_Link_propagatesHeightChanges_twoLargeGraphs(t *testing.T) {
	/*
		This test asserts that height changes propagate beyond
		the first order of children for a given node that is linked.
	*/
	a00 := newMockBareNode()
	a01 := newMockBareNode()

	a10 := newMockBareNode()

	a20 := newMockBareNode()
	a21 := newMockBareNode()

	Link(a10, a00, a01)
	Link(a20, a10)
	Link(a21, a10)

	testutil.ItsEqual(t, 1, a00.Node().height)
	testutil.ItsEqual(t, 1, a01.Node().height)
	testutil.ItsEqual(t, 2, a10.Node().height)
	testutil.ItsEqual(t, 3, a20.Node().height)
	testutil.ItsEqual(t, 3, a21.Node().height)

	b00 := newMockBareNode()
	b01 := newMockBareNode()

	b10 := newMockBareNode()

	b20 := newMockBareNode()
	b21 := newMockBareNode()

	Link(b10, b00, b01)
	Link(b20, b10)
	Link(b21, b10)

	testutil.ItsEqual(t, 1, b00.Node().height)
	testutil.ItsEqual(t, 1, b01.Node().height)
	testutil.ItsEqual(t, 2, b10.Node().height)
	testutil.ItsEqual(t, 3, b20.Node().height)
	testutil.ItsEqual(t, 3, b21.Node().height)

	Link(b00, a20)

	testutil.ItsEqual(t, 4, b00.Node().height)
	testutil.ItsEqual(t, 1, b01.Node().height)
	testutil.ItsEqual(t, 5, b10.Node().height)
	testutil.ItsEqual(t, 6, b20.Node().height)
	testutil.ItsEqual(t, 6, b21.Node().height)
}

func Test_Link_detectsCycles(t *testing.T) {
	n0 := newMockBareNode()
	n1 := newMockBareNode()
	n2 := newMockBareNode()

	err := link(n0, n1)
	testutil.ItsNil(t, err)

	err = link(n1, n2)
	testutil.ItsNil(t, err)

	err = link(n2, n0)
	testutil.ItsNotNil(t, err)
}
