package incr

import (
	"fmt"
)

// Return yields a constant incremental for a given value.
//
// Note that it does not implement `IStabilize` and is effectively
// always the same value (and treated as such).
func Return[A any](scope Scope, v A) Incr[A] {
	return WithinScope(scope, &returnIncr[A]{
		n: NewNode("return"),
		v: v,
	})
}

var (
	_ Incr[string]         = (*returnIncr[string])(nil)
	_ IShouldBeInvalidated = (*returnIncr[string])(nil)
	_ IStale               = (*returnIncr[string])(nil)
	_ fmt.Stringer         = (*returnIncr[string])(nil)
)

type returnIncr[A any] struct {
	n *Node
	v A
}

func (r *returnIncr[A]) Parents() []INode { return nil }

func (r returnIncr[A]) Stale() bool {
	return r.n.recomputedAt == 0
}

func (vn *returnIncr[T]) ShouldBeInvalidated() bool {
	return false
}

func (r returnIncr[A]) Node() *Node { return r.n }

func (r returnIncr[A]) Value() A { return r.v }

func (r returnIncr[A]) String() string { return r.n.String() }
