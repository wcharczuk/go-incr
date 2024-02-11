package incr

import "context"

// Bind3 lets you swap out an entire subgraph of a computation based
// on a given set of 3 input incrementals.
func Bind3[A, B, C, D any](scope Scope, a Incr[A], b Incr[B], c Incr[C], fn func(Scope, A, B, C) Incr[D]) BindIncr[D] {
	return Bind3Context(scope, a, b, c, func(_ context.Context, scope Scope, x0 A, x1 B, x2 C) (Incr[D], error) {
		return fn(scope, x0, x1, x2), nil
	})
}

// Bind3Context lets you swap out an entire subgraph of a computation based
// on a given set of 3 input incrementals, taking a context and as well returning an error.
func Bind3Context[A, B, C, D any](scope Scope, a Incr[A], b Incr[B], c Incr[C], fn func(context.Context, Scope, A, B, C) (Incr[D], error)) BindIncr[D] {
	m := Map3(scope, a, b, c, func(av A, bv B, cv C) tuple3[A, B, C] {
		return tuple3[A, B, C]{av, bv, cv}
	})
	bind := BindContext[tuple3[A, B, C], D](scope, m, func(ctx context.Context, bs Scope, tv tuple3[A, B, C]) (Incr[D], error) {
		return fn(ctx, scope, tv.A, tv.B, tv.C)
	})
	bind.Node().SetKind("bind3")
	return bind
}

// tuple3 is a tuple of values.
type tuple3[A, B, C any] struct {
	A A
	B B
	C C
}
