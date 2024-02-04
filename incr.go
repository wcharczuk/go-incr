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

// IBind implements bind steps for nested actions.
type IBind interface {
	Link(context.Context) error
	Unlink(context.Context)
}

// IUnobserve is a type that may need to implement
// extra steps when it's unobserved.
type IUnobserve interface {
	Unobserve(context.Context)
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
