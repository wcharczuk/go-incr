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
		parentNode := p.getNode()
		parentNode.children = append(parentNode.children, n.self)
		n.parents = append(n.parents, p)
		if !parentNode.changedAt.IsZero() {
			n.changedAt = parentNode.changedAt
		}
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
	return !n.changedAt.IsZero() && n.changedAt.After(n.recomputedAt)
}
