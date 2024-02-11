package incr

// BindIf lets you swap out an entire subgraph of a computation based
// on a given boolean incremental predicate.
func Bind2[A, B, C any](scope Scope, a Incr[A], b Incr[B], fn func(Scope, A, B) Incr[C]) BindIncr[C] {
	m := Map2(scope, a, b, func(av A, bv B) Tuple2[A, B] {
		return Tuple2[A, B]{av, bv}
	})
	bind := Bind[Tuple2[A, B], C](scope, m, func(bs Scope, tv Tuple2[A, B]) Incr[C] {
		return fn(scope, tv.A, tv.B)
	})
	bind.Node().SetKind("bind2")
	return bind
}

// Tuple2 is a tuple of values.
type Tuple2[A, B any] struct {
	A A
	B B
}
