package incr

// BindIf lets you swap out an entire subgraph of a computation based
// on a given boolean incremental predicate.
func Bind4[A, B, C, D, E any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], fn func(Scope, A, B, C, D) Incr[E]) BindIncr[E] {
	m := Map4(scope, a, b, c, d, func(av A, bv B, cv C, dv D) Tuple4[A, B, C, D] {
		return Tuple4[A, B, C, D]{av, bv, cv, dv}
	})
	bind := Bind[Tuple4[A, B, C, D], E](scope, m, func(bs Scope, tv Tuple4[A, B, C, D]) Incr[E] {
		return fn(scope, tv.A, tv.B, tv.C, tv.D)
	})
	bind.Node().SetKind("bind4")
	return bind
}

// Tuple4 is a tuple of values.
type Tuple4[A, B, C, D any] struct {
	A A
	B B
	C C
	D D
}
