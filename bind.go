package incr

import "context"

// Bind lets you swap out an entire subgraph of a computation based
// on a given function and a single input.
func Bind[A, B any](a Incr[A], fn func(A) Incr[B]) Incr[B] {
	o := &bindIncr[A, B]{
		n:  NewNode(),
		a:  a,
		fn: fn,
	}
	Link(o, a)
	return o
}

var (
	_ Incr[bool] = (*bindIncr[string, bool])(nil)
	_ GraphNode  = (*bindIncr[string, string])(nil)
	_ Stabilizer = (*bindIncr[string, string])(nil)
)

type bindIncr[A, B any] struct {
	n     *Node
	a     Incr[A]
	fn    func(A) Incr[B]
	bind  Incr[B]
	value B
}

func (b *bindIncr[A, B]) Node() *Node { return b.n }

func (b *bindIncr[A, B]) Value() B { return b.value }

func (b *bindIncr[A, B]) SetBind(v Incr[B]) {
	b.bind = v
}

func (b *bindIncr[A, B]) Bind() (oldValue, newValue Incr[B]) {
	return b.bind, b.fn(b.a.Value())
}

func (b *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	BindUpdate[B](ctx, b)
	b.value = b.bind.Value()
	return nil
}
