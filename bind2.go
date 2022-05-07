package incr

import (
	"context"
	"fmt"
)

// Bind2 lets you swap out an entire subgraph of a computation based
// on a given function and two inputs.
func Bind2[A, B, C any](a Incr[A], b Incr[B], fn func(context.Context, A, B) (Incr[C], error)) Incr[C] {
	o := &bind2Incr[A, B, C]{
		n:  NewNode(),
		a:  a,
		b:  b,
		fn: fn,
	}
	Link(o, a, b)
	return o
}

var (
	_ Incr[bool]   = (*bind2Incr[string, float64, bool])(nil)
	_ Binder[bool] = (*bind2Incr[string, float64, bool])(nil)
	_ GraphNode    = (*bind2Incr[string, float64, bool])(nil)
	_ Stabilizer   = (*bind2Incr[string, float64, bool])(nil)
	_ fmt.Stringer = (*bind2Incr[string, float64, bool])(nil)
)

type bind2Incr[A, B, C any] struct {
	n     *Node
	a     Incr[A]
	b     Incr[B]
	fn    func(context.Context, A, B) (Incr[C], error)
	bind  Incr[C]
	value C
}

func (b *bind2Incr[A, B, C]) Node() *Node { return b.n }

func (b *bind2Incr[A, B, C]) Value() C { return b.value }

func (b *bind2Incr[A, B, C]) SetBind(v Incr[C]) {
	b.bind = v
}

func (b *bind2Incr[A, B, C]) Bind(ctx context.Context) (oldValue, newValue Incr[C], err error) {
	oldValue = b.bind
	newValue, err = b.fn(ctx, b.a.Value(), b.b.Value())
	return
}

func (b *bind2Incr[A, B, C]) Stabilize(ctx context.Context) error {
	if err := BindUpdate[C](ctx, b); err != nil {
		return err
	}
	b.value = b.bind.Value()
	return nil
}

func (b *bind2Incr[A, B, C]) String() string {
	return "bind2[" + b.n.id.Short() + "]"
}
