package incr

// Incr is the base interface for any incremental node.
type Incr[A comparable] interface {
	Stabilizer
	Value() A
}
