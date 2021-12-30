package incr

// Incr is the base interface for any incremental node.
type Incr[A any] interface {
	ExtIncr[A]
	getNode() *node
}
