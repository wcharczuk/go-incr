package incr

import (
	"context"
)

// Incr is a type that can be an incremental node in a computation graph.
type Incr[T any] interface {
	INode
	// Value returns the stabilized
	// value of the node.
	//
	// Note that you do not want to access this value
	// directly, you almost always want to associate this
	// value with a node as an input, and use the value
	// as resolved through a map or bind function.
	Value() T
}

// INode is a node in the incremental graph.
type INode interface {
	// INode returns the underlying node metadata
	// for a given incremental node.
	Node() *Node
}

// IParents is a type that has parents or inputs.
//
// This list is used to link the node into the graph
// but also to reconstruct the links if the node becomes
// unnecessary, and then necessary again after construction.
type IParents interface {
	// Parents provides a list of nodes that this node takes as inputs
	// for propagation of changes and dependency tracking.
	Parents() []INode
}

// IStabilize is a type that can be stabilized.
type IStabilize interface {
	Stabilize(context.Context) error
}

// IShouldBeInvalidateds a type that determines if a
// node should be invalidated or not.
type IShouldBeInvalidated interface {
	ShouldBeInvalidated() bool
}

// IStale is a type that determines if it's stale or not.
type IStale interface {
	Stale() bool
}

// ICutoff is a type that determines if changes should
// continue to propagate or not.
type ICutoff interface {
	Cutoff(context.Context) (bool, error)
}

// IAlways is a type that determines a node is always marked
// for recomputation.
type IAlways interface {
	Always()
}
