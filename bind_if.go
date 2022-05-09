package incr

import (
	"context"
	"fmt"
)

// BindIf lets you swap out an entire subgraph of a computation based
// on a given boolean incremental predicate.
func BindIf[A any](p Incr[bool], fn func(context.Context, bool) (Incr[A], error)) Incr[A] {
	o := &bindIfIncr[A]{
		n:  NewNode(),
		fn: fn,
		p:  p,
	}
	// NOTE(wc): a | b will be linked when this stabilizes
	// but the predicate is _always_ linked and part of the graph
	Link(o, p)
	return o
}

var (
	_ Incr[string]  = (*bindIfIncr[string])(nil)
	_ IBind[string] = (*bindIfIncr[string])(nil)
	_ INode         = (*bindIfIncr[string])(nil)
	_ IStabilize    = (*bindIfIncr[string])(nil)
	_ fmt.Stringer  = (*bindIfIncr[string])(nil)
)

type bindIfIncr[A any] struct {
	n     *Node
	p     Incr[bool]
	fn    func(context.Context, bool) (Incr[A], error)
	bound Incr[A]
	value A
}

func (b *bindIfIncr[A]) Node() *Node { return b.n }

func (b *bindIfIncr[A]) Value() A { return b.value }

func (b *bindIfIncr[A]) SetBind(v Incr[A]) {
	b.bound = v
}

func (b *bindIfIncr[A]) Bind(ctx context.Context) (oldValue, newValue Incr[A], err error) {
	oldValue = b.bound
	newValue, err = b.fn(ctx, b.p.Value())
	return
}

func (b *bindIfIncr[A]) Stabilize(ctx context.Context) error {
	if err := bindUpdate[A](ctx, b); err != nil {
		return err
	}
	b.value = b.bound.Value()
	return nil
}

func (b *bindIfIncr[A]) String() string {
	return FormatNode(b.n, "bind_if")
}
