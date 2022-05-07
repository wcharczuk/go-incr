package incr

import "context"

// Bind lets you swap out an entire subgraph of a computation based
// on a given function.
//
// Bind is really important in filtering large graphs of computations
// where a Map node would recompute if _any_ of the inputs change,
// a Bind node would only recompute if the current selection changes.
func Bind[A, B any](a Incr[A], fn func(A) Incr[B]) Incr[B] {
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
func BindUpdate[A any](ctx context.Context, b Binder[A]) {
	oldValue, newValue := b.Bind()
	if oldValue == nil {
		Link(newValue, b)
		discoverAllNodes(ctx, b.Node().gs, newValue)
		b.SetBind(newValue)
		return
	}

	if oldValue.Node().id != newValue.Node().id {
		// purge old value and all parents from recompute heap
		Unlink(oldValue)
		undiscoverAllNodes(ctx, b.Node().gs, oldValue)

		Link(newValue, b)
		discoverAllNodes(ctx, b.Node().gs, newValue)
		b.SetBind(newValue)
	}
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
