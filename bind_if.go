package incr

import (
	"context"
	"fmt"
)

// BindIf lets you swap out an entire subgraph of a computation based
// on a given boolean incremental predicate.
func BindIf[A any](a Incr[A], b Incr[A], p Incr[bool]) Incr[A] {
	o := &bindIfIncr[A]{
		n: NewNode(),
		a: a,
		b: b,
		p: p,
	}
	// NOTE(wc): a | b will be linked when this stabilizes
	Link(o, p)
	return o
}

var (
	_ Incr[string]   = (*bindIfIncr[string])(nil)
	_ GraphNode      = (*bindIfIncr[string])(nil)
	_ Binder[string] = (*bindIfIncr[string])(nil)
	_ Stabilizer     = (*bindIfIncr[string])(nil)
	_ fmt.Stringer   = (*bindIfIncr[string])(nil)
)

type bindIfIncr[A any] struct {
	n     *Node
	a     Incr[A]
	b     Incr[A]
	p     Incr[bool]
	bind  Incr[A]
	value A
}

func (b *bindIfIncr[A]) Node() *Node { return b.n }

func (b *bindIfIncr[A]) Value() A { return b.value }

func (b *bindIfIncr[A]) SetBind(v Incr[A]) {
	b.bind = v
}

func (b *bindIfIncr[A]) Bind(_ context.Context) (oldValue, newValue Incr[A], err error) {
	if b.p.Value() {
		return b.bind, b.a, nil
	}
	return b.bind, b.b, nil
}

func (b *bindIfIncr[A]) Stabilize(ctx context.Context) error {
	if err := BindUpdate[A](ctx, b); err != nil {
		return err
	}
	b.value = b.bind.Value()
	return nil
}

func (b *bindIfIncr[A]) String() string {
	return "bind_if[" + b.n.id.Short() + "]"
}
