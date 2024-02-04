package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ExpertNode_setters(t *testing.T) {
	n := newMockBareNode()
	en := ExpertNode(n)

	id := NewIdentifier()
	en.SetID(id)

	testutil.ItsEqual(t, id, n.Node().ID())

	en.SetHeight(7)
	testutil.ItsEqual(t, 7, en.Height())

	en.SetChangedAt(8)
	testutil.ItsEqual(t, 8, en.ChangedAt())

	en.SetSetAt(9)
	testutil.ItsEqual(t, 9, en.SetAt())

	en.SetRecomputedAt(10)
	testutil.ItsEqual(t, 10, en.RecomputedAt())

	en.SetBoundAt(11)
	testutil.ItsEqual(t, 11, en.BoundAt())

	en.SetAlways(true)
	testutil.ItsEqual(t, true, en.Always())
}

func Test_ExpertNode_AddChildren(t *testing.T) {
	n := newMockBareNode()
	en := ExpertNode(n)

	en.AddChildren(newMockBareNode(), newMockBareNode())
	testutil.ItsEqual(t, 2, len(n.Node().Children()))
}

func Test_ExpertNode_AddParents(t *testing.T) {
	n := newMockBareNode()
	en := ExpertNode(n)

	en.AddParents(newMockBareNode(), newMockBareNode())
	testutil.ItsEqual(t, 2, len(n.Node().Parents()))
}

func Test_ExpertNode_RemoveChild(t *testing.T) {
	n := newMockBareNode()
	en := ExpertNode(n)

	mbn0 := newMockBareNode()
	mbn1 := newMockBareNode()
	en.AddChildren(mbn0, mbn1)
	testutil.ItsEqual(t, 2, len(n.Node().Children()))

	en.RemoveChild(mbn0.Node().ID())
	testutil.ItsEqual(t, 1, len(n.Node().Children()))
}

func Test_ExpertNode_RemoveParent(t *testing.T) {
	n := newMockBareNode()
	en := ExpertNode(n)

	mbn0 := newMockBareNode()
	mbn1 := newMockBareNode()
	en.AddParents(mbn0, mbn1)
	testutil.ItsEqual(t, 2, len(n.Node().Parents()))

	en.RemoveParent(mbn0.Node().ID())
	testutil.ItsEqual(t, 1, len(n.Node().Parents()))
}

func Test_ExpertNode_Value(t *testing.T) {
	n := Return(testContext(), "hello")
	en := ExpertNode(n)

	value := en.Value()
	testutil.ItsEqual(t, "hello", value)
}

func Test_ExperNode_ComputePseudoHeight(t *testing.T) {
	a00 := newMockBareNode()
	a01 := newMockBareNode()
	a10 := newMockBareNode()
	a20 := newMockBareNode()
	a21 := newMockBareNode()

	Link(a10, a00, a01)
	Link(a20, a10)
	Link(a21, a10)

	b00 := newMockBareNode()
	b01 := newMockBareNode()
	b10 := newMockBareNode()
	b20 := newMockBareNode()
	b21 := newMockBareNode()

	Link(b10, b00, b01)
	Link(b20, b10)
	Link(b21, b10)
	Link(b00, a20)

	testutil.ItsEqual(t, 6, ExpertNode(b21).ComputePseudoHeight())
}
