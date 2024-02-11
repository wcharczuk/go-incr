package incr

import "context"

// Bind4 lets you swap out an entire subgraph of a computation based
// on a given set of 4 input incrementals.
func Bind4[A, B, C, D, E any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], fn func(Scope, A, B, C, D) Incr[E]) BindIncr[E] {
	return Bind4Context(scope, a, b, c, d, func(_ context.Context, scope Scope, x0 A, x1 B, x2 C, x3 D) (Incr[E], error) {
		return fn(scope, x0, x1, x2, x3), nil
	})
}

// Bind4Context lets you swap out an entire subgraph of a computation based
// on a given set of 4 input incrementals, taking a context and as well returning an error.
func Bind4Context[A, B, C, D, E any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], fn func(context.Context, Scope, A, B, C, D) (Incr[E], error)) BindIncr[E] {
	m := Map4(scope, a, b, c, d, func(av A, bv B, cv C, dv D) tuple4[A, B, C, D] {
		return tuple4[A, B, C, D]{av, bv, cv, dv}
	})
	bind := BindContext[tuple4[A, B, C, D], E](scope, m, func(ctx context.Context, bs Scope, tv tuple4[A, B, C, D]) (Incr[E], error) {
		return fn(ctx, scope, tv.A, tv.B, tv.C, tv.D)
	})
	bind.Node().SetKind("bind4")
	return bind
}

// tuple4 is a tuple of values.
type tuple4[A, B, C, D any] struct {
	A A
	B B
	C C
	D D
}
