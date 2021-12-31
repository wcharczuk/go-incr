package incr

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
		n.height = parentNode.height + 1
	}
}

type generation uint64

type node struct {
	id nodeID

	height int

	self Stabilizer

	parents  []Stabilizer
	children []Stabilizer
}
