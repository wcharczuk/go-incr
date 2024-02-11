package incr

// BindIf lets you swap out an entire subgraph of a computation based
// on a given boolean incremental predicate.
func Bind3[A, B, C, D any](scope Scope, a Incr[A], b Incr[B], c Incr[C], fn func(Scope, A, B, C) Incr[D]) BindIncr[D] {
	m := Map3(scope, a, b, c, func(av A, bv B, cv C) Tuple3[A, B, C] {
		return Tuple3[A, B, C]{av, bv, cv}
	})
	bind := Bind[Tuple3[A, B, C], D](scope, m, func(bs Scope, tv Tuple3[A, B, C]) Incr[D] {
		return fn(scope, tv.A, tv.B, tv.C)
	})
	bind.Node().SetKind("bind3")
	return bind
}

// Tuple3 is a tuple of values.
type Tuple3[A, B, C any] struct {
	A A
	B B
	C C
}
