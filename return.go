package incr

import (
	"context"
	"fmt"
)

// Return yields a constant incremental for a given value.
//
// Note that it does not implement `IStabilize` and is effectively
// always the same value (and treated as such).
func Return[A any](ctx context.Context, v A) Incr[A] {
	return WithBindScope(ctx, &returnIncr[A]{
		n: NewNode(),
		v: v,
	})
}

var (
	_ Incr[string] = (*returnIncr[string])(nil)
	_ INode        = (*returnIncr[string])(nil)
	_ fmt.Stringer = (*returnIncr[string])(nil)
)

type returnIncr[A any] struct {
	n *Node
	v A
}

func (r returnIncr[A]) Node() *Node { return r.n }

func (r returnIncr[A]) Value() A { return r.v }

func (r returnIncr[A]) String() string { return r.n.String("return") }
