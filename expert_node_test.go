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

	testutil.Equal(t, id, n.Node().ID())

	en.SetHeight(7)
	testutil.Equal(t, 7, en.Height())

	en.SetChangedAt(8)
	testutil.Equal(t, 8, en.ChangedAt())

	en.SetSetAt(9)
	testutil.Equal(t, 9, en.SetAt())

	en.SetRecomputedAt(10)
	testutil.Equal(t, 10, en.RecomputedAt())

	en.SetBoundAt(11)
	testutil.Equal(t, 11, en.BoundAt())

	en.SetAlways(true)
	testutil.Equal(t, true, en.Always())
}

func Test_ExpertNode_AddChildren(t *testing.T) {
	n := newMockBareNode()
	en := ExpertNode(n)

	en.AddChildren(newMockBareNode(), newMockBareNode())
	testutil.Equal(t, 2, len(n.Node().Children()))
}

func Test_ExpertNode_AddParents(t *testing.T) {
	n := newMockBareNode()
	en := ExpertNode(n)

	en.AddParents(newMockBareNode(), newMockBareNode())
	testutil.Equal(t, 2, len(n.Node().Parents()))
}

func Test_ExpertNode_RemoveChild(t *testing.T) {
	n := newMockBareNode()
	en := ExpertNode(n)

	mbn0 := newMockBareNode()
	mbn1 := newMockBareNode()
	en.AddChildren(mbn0, mbn1)
	testutil.Equal(t, 2, len(n.Node().Children()))

	en.RemoveChild(mbn0.Node().ID())
	testutil.Equal(t, 1, len(n.Node().Children()))
}

func Test_ExpertNode_RemoveParent(t *testing.T) {
	n := newMockBareNode()
	en := ExpertNode(n)

	mbn0 := newMockBareNode()
	mbn1 := newMockBareNode()
	en.AddParents(mbn0, mbn1)
	testutil.Equal(t, 2, len(n.Node().Parents()))

	en.RemoveParent(mbn0.Node().ID())
	testutil.Equal(t, 1, len(n.Node().Parents()))
}

func Test_ExpertNode_Value(t *testing.T) {
	g := New()
	n := Return(g, "hello")
	en := ExpertNode(n)

	value := en.Value()
	testutil.Equal(t, "hello", value)
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

	testutil.Equal(t, 6, ExpertNode(b21).ComputePseudoHeight())
}

func Test_ExperNode_ComputePseudoHeight_bare(t *testing.T) {
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

	ExpertNode(a00).SetHeight(0)
	ExpertNode(a01).SetHeight(0)
	ExpertNode(a10).SetHeight(0)
	ExpertNode(a20).SetHeight(0)
	ExpertNode(a21).SetHeight(0)
	ExpertNode(b00).SetHeight(0)
	ExpertNode(b01).SetHeight(0)
	ExpertNode(b10).SetHeight(0)
	ExpertNode(b20).SetHeight(0)
	ExpertNode(b21).SetHeight(0)

	testutil.Equal(t, 6, ExpertNode(b21).ComputePseudoHeight())
}
