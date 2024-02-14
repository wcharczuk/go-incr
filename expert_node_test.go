package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_ExpertNode_setters(t *testing.T) {
	g := New()
	n := newMockBareNode(g)
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

	en.SetAlways(true)
	testutil.Equal(t, true, en.Always())
}

func Test_ExpertNode_AddChildren(t *testing.T) {
	g := New()
	n := newMockBareNode(g)
	en := ExpertNode(n)

	en.AddChildren(newMockBareNode(g), newMockBareNode(g))
	testutil.Equal(t, 2, len(n.Node().Children()))
}

func Test_ExpertNode_AddParents(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	en := ExpertNode(n)

	en.AddParents(newMockBareNode(g), newMockBareNode(g))
	testutil.Equal(t, 2, len(n.Node().Parents()))
}

func Test_ExpertNode_RemoveChild(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	en := ExpertNode(n)

	mbn0 := newMockBareNode(g)
	mbn1 := newMockBareNode(g)
	en.AddChildren(mbn0, mbn1)
	testutil.Equal(t, 2, len(n.Node().Children()))

	en.RemoveChild(mbn0.Node().ID())
	testutil.Equal(t, 1, len(n.Node().Children()))
}

func Test_ExpertNode_RemoveParent(t *testing.T) {
	g := New()

	n := newMockBareNode(g)
	en := ExpertNode(n)

	mbn0 := newMockBareNode(g)
	mbn1 := newMockBareNode(g)
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
