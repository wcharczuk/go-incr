package incr

import (
	"context"
	"fmt"
)

// Bind lets you swap out an entire subgraph of a computation based
// on a given function and a single input.
func Bind[A, B any](a Incr[A], fn func(context.Context, A) (Incr[B], error)) Incr[B] {
	o := &bindIncr[A, B]{
		n:  NewNode(),
		a:  a,
		fn: fn,
	}
	Link(o, a)
	return o
}

var (
	_ Incr[bool]   = (*bindIncr[string, bool])(nil)
	_ Binder[bool] = (*bindIncr[string, bool])(nil)
	_ GraphNode    = (*bindIncr[string, bool])(nil)
	_ Stabilizer   = (*bindIncr[string, bool])(nil)
	_ fmt.Stringer = (*bindIncr[string, bool])(nil)
)

type bindIncr[A, B any] struct {
	n     *Node
	a     Incr[A]
	fn    func(context.Context, A) (Incr[B], error)
	bind  Incr[B]
	value B
}

func (b *bindIncr[A, B]) Node() *Node { return b.n }

func (b *bindIncr[A, B]) Value() B { return b.value }

func (b *bindIncr[A, B]) SetBind(v Incr[B]) {
	b.bind = v
}

func (b *bindIncr[A, B]) Bind(ctx context.Context) (oldValue, newValue Incr[B], err error) {
	oldValue = b.bind
	newValue, err = b.fn(ctx, b.a.Value())
	return
}

func (b *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	if err := BindUpdate[B](ctx, b); err != nil {
		return err
	}
	b.value = b.bind.Value()
	return nil
}

func (b *bindIncr[A, B]) String() string {
	return "bind[" + b.n.id.Short() + "]"
}
