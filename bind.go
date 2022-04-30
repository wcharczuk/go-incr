package incr

// Bind lets you swap out an entire subgraph of a computation based
// on a given function.
func Bind[A, B comparable](a Incr[A], fn func(A) Incr[B]) Incr[B] {
	return nil
}
