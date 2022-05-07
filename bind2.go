package incr

import "context"

// Bind2 lets you swap out an entire subgraph of a computation based
// on a given function and two inputs.
func Bind2[A, B, C any](a Incr[A], b Incr[B], fn func(A, B) Incr[C]) Incr[C] {
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
	_ Incr[bool] = (*bind2Incr[string, float64, bool])(nil)
	_ GraphNode  = (*bind2Incr[string, float64, string])(nil)
	_ Stabilizer = (*bind2Incr[string, float64, string])(nil)
)

type bind2Incr[A, B, C any] struct {
	n     *Node
	a     Incr[A]
	b     Incr[B]
	fn    func(A, B) Incr[C]
	bind2 Incr[C]
	value C
}

func (b *bind2Incr[A, B, C]) Node() *Node { return b.n }

func (b *bind2Incr[A, B, C]) Value() C { return b.value }

func (b *bind2Incr[A, B, C]) SetBind(v Incr[C]) {
	b.bind2 = v
}

func (b *bind2Incr[A, B, C]) Bind() (oldValue, newValue Incr[C]) {
	return b.bind2, b.fn(b.a.Value(), b.b.Value())
}

func (b *bind2Incr[A, B, C]) Stabilize(ctx context.Context) error {
	BindUpdate[C](ctx, b)
	b.value = b.bind2.Value()
	return nil
}