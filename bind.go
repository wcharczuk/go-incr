package incr

import (
	"context"
)

// Bind returns the result of a given function `fn` on a given input.
func Bind[A, B any](i Incr[A], fn func(A) Incr[B]) Incr[B] {
	b := &bindIncr[A, B]{
		i:  i,
		fn: fn,
	}
	b.node = newNode(
		b,
		optNodeChildOf(i),
	)
	return b
}

// bindIncr is a concrete implementation of Incr for
// the bind operator.
type bindIncr[A, B any] struct {
	*node
	i  Incr[A]
	fn func(A) Incr[B]
}

// Value implements Incr[B]
func (bi bindIncr[A, B]) Value() B {
	return bi.fn(bi.i.Value()).Value()
}

// Stabilize implements Incr[B]
func (bi bindIncr[A, B]) Stabilize(ctx context.Context) error {
	return nil
}

func (bi bindIncr[A, B]) getNode() *node {
	return bi.node
}
