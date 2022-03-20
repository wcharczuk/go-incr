package incr

import (
	"context"
	"time"
)

// NewNode returns a new node.
func NewNode(self Stabilizer, opts ...NodeOption) *Node {
	n := &Node{
		id:   NewNodeID(),
		self: self,
		now:  time.Now,
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// ReplaceNode replaces a node.
//
// This is typically used in Bind and nowhere else.
func ReplaceNode(old *Node, opts ...NodeOption) *Node {
	n := &Node{
		id:   old.id,
		self: old.self,
		now:  old.now,
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

// NodeOption mutates a node.
type NodeOption func(*Node)

// OptNodeChildOf sets the node to be the child of another node.
func OptNodeChildOf(p Stabilizer) NodeOption {
	return func(n *Node) {
		parentNode := p.Node()
		parentNode.children = append(parentNode.children, n.self)
		n.height = Max(n.height, parentNode.height+1)
		n.parents = append(n.parents, p)
	}
}

type Node struct {
	id     NodeID
	height int

	initialized  bool
	recomputedAt time.Time

	self     Stabilizer
	parents  []Stabilizer
	children []Stabilizer

	now      func() time.Time
	onUpdate []func(context.Context)
}

// OnUpdate registers on update handlers that are called
// when the node is stabilized.
func (n *Node) OnUpdate(handler func(context.Context)) {
	n.onUpdate = append(n.onUpdate, handler)
}
