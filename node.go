package incr

import "time"

func newNode(self Stabilizer, opts ...nodeOption) *node {
	n := &node{
		id:   newNodeID(),
		self: self,
	}
	for _, opt := range opts {
		opt(n)
	}
	return n
}

type nodeOption func(*node)

func optNodeChildOf(p Stabilizer) nodeOption {
	return func(n *node) {
		p.getNode().children = append(p.getNode().children, n.self)
		n.parents = append(n.parents, p)
	}
}

type node struct {
	id nodeID

	self Stabilizer

	recomputedAt time.Time
	changedAt    time.Time

	parents  []Stabilizer
	children []Stabilizer
}

// isStale returns if the node is stale.
func (n *node) isStale() bool {
	if n.changedAt.IsZero() {
		return false
	}
	return n.changedAt.After(n.recomputedAt)
}

func (n *node) setChangedAtRecursive(changedAt time.Time) {
	n.changedAt = changedAt
	for _, child := range n.children {
		child.getNode().setChangedAtRecursive(changedAt)
	}
}
