package incr

import "context"

// ExtIncr is a non-package type that can be used as an incr node.
type ExtIncr[A any] interface {
	Staler
	Value() A
	Stabilize(context.Context) error
}
