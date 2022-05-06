package incr

func newNode() *Node {
	return &Node{id: newNodeID()}
}

// Node is the common metadata for any node in the computation graph.
type Node struct {
	// id is a unique identifier for the node.
	id nodeID
	// gs is a shared reference to the graph state
	// for the computation
	gs *graphState
	// parents are the nodes that this node
	// depends on, and would require this node
	// to be recomputed if they changd.
	parents []Stabilizer
	// children are the nodes that depend on this
	// node, and if this node chanages, would need
	// to be recomputed.
	children      []Stabilizer
	status        Status
	height        int
	initializedAt uint64
	changedAt     uint64
	recomputedAt  uint64
}
