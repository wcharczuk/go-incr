package incr

import "context"

// BindIf lets you swap out an entire subgraph of a computation based
// on a given boolean incremental predicate.
//
// It is largely "macro" for a Bind that takes an input bool incremental.
func BindIf[A any](scope Scope, p Incr[bool], fn func(context.Context, Scope, bool) (Incr[A], error)) BindIncr[A] {
	b := BindContext(scope, p, fn)
	b.Node().SetKind("bind_if")
	return b
}
