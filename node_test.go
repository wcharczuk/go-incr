package incr

import "testing"

func Test_Node_NewNode(t *testing.T) {
	n := NewNode()
	ItsNotNil(t, n.id)
	ItsNil(t, n.gs)
	ItsNil(t, n.parents)
	ItsNil(t, n.children)
	ItsEqual(t, "", n.label)
	ItsEqual(t, 0, n.height)
	ItsEqual(t, 0, n.changedAt)
	ItsEqual(t, 0, n.setAt)
	ItsEqual(t, 0, n.recomputedAt)
	ItsNil(t, n.onUpdateHandlers)
	ItsNil(t, n.onErrorHandlers)
	ItsNil(t, n.stabilize)
	ItsNil(t, n.cutoff)
	ItsEqual(t, 0, n.numRecomputes)
}

type mockNode struct {
	n *Node
}

func (mn *mockNode) Node() *Node {
	if mn.n == nil {
		mn.n = NewNode()
	}
	return mn.n
}

func Test_Link(t *testing.T) {
	p := new(mockNode)
	c0 := new(mockNode)
	c1 := new(mockNode)
	c2 := new(mockNode)

	// set up P with (3) inputs
	Link(p, c0, c1, c2)

	// no nodes depend on p, p is not an input to any nodes
	ItsEqual(t, 0, len(p.n.parents))
	ItsEqual(t, 3, len(p.n.children))
	ItsEqual(t, c0.n.id, p.n.children[0].Node().id)
	ItsEqual(t, c1.n.id, p.n.children[1].Node().id)
	ItsEqual(t, c2.n.id, p.n.children[2].Node().id)

	ItsEqual(t, 1, len(c0.n.parents))
	ItsEqual(t, p.n.id, c0.n.parents[0].Node().id)
	ItsEqual(t, 0, len(c0.n.children))

	ItsEqual(t, 1, len(c1.n.parents))
	ItsEqual(t, p.n.id, c1.n.parents[0].Node().id)
	ItsEqual(t, 0, len(c1.n.children))

	ItsEqual(t, 1, len(c2.n.parents))
	ItsEqual(t, p.n.id, c2.n.parents[0].Node().id)
	ItsEqual(t, 0, len(c2.n.children))
}
