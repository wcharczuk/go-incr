package incr

import (
	"context"
)

// Bind returns the result of a given function `fn` on a given input.
//
// It differs from `Map` in that the provided function must return an Incr[B]
// as opposed to `Map` that returns just a B.
//
// The implication of returning an Incr[B] is that Bind can replace itself
// and the
func Bind[A, B comparable](i Incr[A], fn func(A) Incr[B]) BindIncr[B] {
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

// BindIncr is the interface a Bind implements
type BindIncr[A comparable] interface {
	Incr[A]
	Incr() Incr[A]
}

type bindIncr[A, B comparable] struct {
	n     *node
	i     Incr[A]
	fn    func(A) Incr[B]
	value Incr[B]
}

func (bi *bindIncr[A, B]) Incr() Incr[B] {
	return bi.value
}

func (bi *bindIncr[A, B]) Value() B {
	return bi.value.Value()
}

func (bi *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	if err := bi.i.Stabilize(ctx); err != nil {
		return err
	}
	bi.value = bi.fn(bi.i.Value())
	return nil
}

func (bi *bindIncr[A, B]) getValue() any {
	return bi.value.Value()
}

func (bi *bindIncr[A, B]) getNode() *node {
	return bi.n
}
