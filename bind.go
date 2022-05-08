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

// BindUpdate is a helper for dealing with bind node changes
// specifically handling unlinking and linking bound nodes
// when the bind changes.
func BindUpdate[A any](ctx context.Context, b IBind[A]) error {
	oldValue, newValue, err := b.Bind(ctx)
	if err != nil {
		return err
	}
	if oldValue == nil {
		Link(newValue, b)
		discoverAllNodes(ctx, b.Node().gs, newValue)
		b.SetBind(newValue)
		return nil
	}

	if oldValue.Node().id != newValue.Node().id {
		Unlink(oldValue)
		undiscoverAllNodes(ctx, b.Node().gs, oldValue)
		Link(newValue, b)
		discoverAllNodes(ctx, b.Node().gs, newValue)
		b.SetBind(newValue)
	}
	return nil
}

var (
	_ Incr[bool]   = (*bindIncr[string, bool])(nil)
	_ IBind[bool]  = (*bindIncr[string, bool])(nil)
	_ INode        = (*bindIncr[string, bool])(nil)
	_ IStabilize   = (*bindIncr[string, bool])(nil)
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
	return FormatNode(b.n, "bind")
}
