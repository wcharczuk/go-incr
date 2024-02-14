package incr

import "context"

// BindIf lets you swap out an entire subgraph of a computation based
// on a given boolean incremental predicate.
func BindIf[A any](scope Scope, p Incr[bool], fn func(context.Context, Scope, bool) (Incr[A], error)) BindIncr[A] {
	b := BindContext[bool, A](scope, p, fn)
	b.Node().SetKind("bind_if")
	return b
}
