package incr

// Join flattens an incremental whose value is itself an incremental.
//
// This is [Bind] with an identity function, which is worth naming because the shape
// comes up on its own: a computation that selects which other computation to read
// from produces an Incr[Incr[A]], and Join is how it is read.
//
// The inner incremental is made necessary while it is the value of the outer one, and
// released when the outer value moves to a different inner incremental, exactly as a
// bind releases a right-hand side it has replaced.
func Join[A any](scope Scope, input Incr[Incr[A]]) Incr[A] {
	return Bind(scope, input, func(_ Scope, inner Incr[A]) Incr[A] {
		return inner
	})
}
