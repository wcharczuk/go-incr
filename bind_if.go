package incr

import (
	"context"
)

// BindIf lets you swap out an entire subgraph of a computation based
// on a given boolean incremental predicate.
func BindIf[A any](ctx context.Context, p Incr[bool], fn func(context.Context, bool) (Incr[A], error)) BindIncr[A] {
	b := BindContext[bool, A](ctx, p, fn).(*bindIncr[bool, A])
	b.bt = "bind_if"
	return WithinBindScope(ctx, b)
}
