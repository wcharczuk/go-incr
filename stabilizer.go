package incr

import "context"

// Stabilizer is a type that can be stabilized.
type Stabilizer interface {
	Staler
	Stabilize(context.Context) error
	getNode() *node
}
