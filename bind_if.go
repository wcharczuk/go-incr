package incr

import "context"

// BindIf lets you swap out an entire subgraph of a computation based
// on a given boolean incremental predicate.
func BindIf[A any](scope *BindScope, p Incr[bool], fn func(context.Context, *BindScope, bool) (Incr[A], error)) BindIncr[A] {
	b := BindContext[bool, A](scope, p, fn).(*bindIncr[bool, A])
	b.bt = "bind_if"
	return WithinBindScope(scope, b)
}
