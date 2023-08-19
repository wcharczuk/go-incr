package incr

import (
	"context"
)

// Incr is a type that can be an incremental node in a computation graph.
type Incr[T any] interface {
	INode
	Value() T
}

// INode is a node in the incremental graph.
type INode interface {
	Node() *Node
}

// IStabilize is a type that can be stabilized.
type IStabilize interface {
	Stabilize(context.Context) error
}

// IBind is a type that can be bound.
type IBind interface {
	Bind(context.Context) error
}

// ICutoff is a type that determines if changes should
// continue to propagate or not.
type ICutoff interface {
	Cutoff(context.Context) bool
}
