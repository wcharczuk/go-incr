package incr

import (
	"context"
)

// Bind2 returns the result of a given function `fn` on given inputs.
//
// It differs from `Map2` in that the provided function must return an Incr[C]
// as opposed to `Map2` that returns just a C.
//
// The implication of returning an Incr[C] is that Bind2 _always_ is stale,
// and will cause the children of any Bind2 node to recompute each pass.
func Bind2[A, B, C comparable](i0 Incr[A], i1 Incr[B], fn func(A, B) Incr[C]) Incr[C] {
	b := &bind2Incr[A, B, C]{
		i0: i0,
		i1: i1,
		fn: fn,
	}
	b.n = newNode(
		b,
		optNodeChildOf(i0),
		optNodeChildOf(i1),
	)
	return b
}

type bind2Incr[A, B, C comparable] struct {
	n     *node
	i0    Incr[A]
	i1    Incr[B]
	fn    func(A, B) Incr[C]
	value Incr[C]
}

func (bi *bind2Incr[A, B, C]) Value() C {
	return bi.value.Value()
}

func (bi *bind2Incr[A, B, C]) Stabilize(ctx context.Context) error {
	bi.value = bi.fn(bi.i0.Value(), bi.i1.Value())
	return nil
}

func (bi *bind2Incr[A, B, C]) getValue() any {
	return bi.Value()
}

func (bi *bind2Incr[A, B, C]) getNode() *node {
	return bi.n
}
