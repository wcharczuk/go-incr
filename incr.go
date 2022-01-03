package incr

import "context"

// Incr is the base interface for any incremental node.
type Incr[A any] interface {
	Stabilizer
	Value() A
}

// Stabilizer is a type that can be stabilized.
type Stabilizer interface {
	Stabilize(context.Context) error
	Stale() bool
	getNode() *node
}
