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
	getNode() *node
}

// Staler lets a node opt out of stabilization.
type Staler interface {
	Stale() bool
}
