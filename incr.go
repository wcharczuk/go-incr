package incr

import (
	"context"
	"fmt"
)

// Incr is a type that can be an incremental node in a computation graph.
type Incr[T any] interface {
	GraphNode
	Value() T
}

type GraphNode interface {
	fmt.Stringer
	Node() *Node
}

// Stabilizer is a type that can be stabilized.
type Stabilizer interface {
	Stabilize(context.Context) error
}

// Cutoffer is a type that determines if changes should
// continue to propagate or not.
type Cutoffer interface {
	Cutoff(context.Context) bool
}
