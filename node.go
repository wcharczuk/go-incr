package incr

// Node is the common metadata for any node in the computation graph.
type Node struct {
	id           NodeID
	parents      []*Node
	status       Status
	height       int
	changedAt    int64
	recomputedAt int64
}
