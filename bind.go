package incr

import (
	"context"
)

// Bind returns the result of a given function `fn` on a given input.
//
// It differs from `Map` in that the provided function must return an Incr[B]
// as opposed to `Map` that returns just a B.
//
// The implication of returning an Incr[B] is that Bind _always_ is stale,
// and will cause the children of any Bind node to recompute each pass.
func Bind[A, B comparable](i Incr[A], fn func(A) Incr[B]) Incr[B] {
	b := &bindIncr[A, B]{
		i:  i,
		fn: fn,
	}
	b.n = newNode(
		b,
		optNodeChildOf(i),
	)
	return b
}

type bindIncr[A, B comparable] struct {
	n     *node
	i     Incr[A]
	fn    func(A) Incr[B]
	value Incr[B]
}

func (bi *bindIncr[A, B]) Value() B {
	return bi.value.Value()
}

func (bi *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	bi.value = bi.fn(bi.i.Value())
	return nil
}

func (bi *bindIncr[A, B]) getValue() any {
	return bi.Value()
}

func (bi *bindIncr[A, B]) getNode() *node {
	return bi.n
}
