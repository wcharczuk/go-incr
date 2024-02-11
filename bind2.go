package incr

import "context"

// Bind2 lets you swap out an entire subgraph of a computation based
// on a given set of 2 input incrementals.
func Bind2[A, B, C any](scope Scope, a Incr[A], b Incr[B], fn func(Scope, A, B) Incr[C]) BindIncr[C] {
	return Bind2Context(scope, a, b, func(_ context.Context, scope Scope, x0 A, x1 B) (Incr[C], error) {
		return fn(scope, x0, x1), nil
	})
}

// Bind2Context lets you swap out an entire subgraph of a computation based
// on a given set of 2 input incrementals, taking a context and as well returning an error.
func Bind2Context[A, B, C any](scope Scope, a Incr[A], b Incr[B], fn func(context.Context, Scope, A, B) (Incr[C], error)) BindIncr[C] {
	m := Map2(scope, a, b, func(av A, bv B) tuple2[A, B] {
		return tuple2[A, B]{av, bv}
	})
	bind := BindContext[tuple2[A, B], C](scope, m, func(ctx context.Context, bs Scope, tv tuple2[A, B]) (Incr[C], error) {
		return fn(ctx, scope, tv.A, tv.B)
	})
	bind.Node().SetKind("bind2")
	return bind
}

// tuple2 is a tuple of values.
type tuple2[A, B any] struct {
	A A
	B B
}
