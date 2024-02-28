package naive

func Bind[A, B any](input Node[A], fn BindFn[A, B]) Node[B] {
	return &bindNodeImpl[A, B]{
		input:  input,
		action: fn,
	}
}

type BindFn[A, B any] func(A) Node[B]

type bindNodeImpl[A, B any] struct {
	input  Node[A]
	action BindFn[A, B]
	bound  Node[B]
}

func (n *bindNodeImpl[A, B]) Value() B {
	n.bound = n.action(n.input.Value())
	return n.bound.Value()
}
