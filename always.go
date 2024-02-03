package incr

import "context"

// Always returns an incremental that is always stale and will be
// marked for recomputation.
func Always[A any](ctx context.Context, input Incr[A]) Incr[A] {
	a := &alwaysIncr[A]{
		n:     NewNode(),
		input: input,
	}
	Link(a, input)
	return WithinBindScope(ctx, a)
}

// AlwaysIncr is a type that implements the always stale incremental.
type AlwaysIncr[A any] interface {
	Incr[A]
	IAlways
}

type alwaysIncr[A any] struct {
	n     *Node
	input Incr[A]
}

func (a *alwaysIncr[A]) Always() {}

func (a *alwaysIncr[A]) Value() A {
	return a.input.Value()
}

func (a *alwaysIncr[A]) Node() *Node { return a.n }

func (a *alwaysIncr[A]) String() string {
	return a.n.String("always")
}
