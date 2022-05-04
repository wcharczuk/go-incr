package incr

// Incr is a type that can be an incremental node in a computation graph.
type Incr[T any] interface {
	Stabilizer
	Value() T
}
