package incr

import (
	"context"
)

// Bind returns the result of a given function `fn` on a given input.
//
// It differs from `Map` in that the provided function must return an Incr[B]
// as opposed to `Map` that returns a value B.
//
// The implication of returning an Incr[B] is that Bind can replace itself
// and the computation below it completely.
func Bind[A, B any](i Incr[A], fn func(A) Incr[B]) BindIncr[B] {
	b := &bindIncr[A, B]{
		i:  i,
		fn: fn,
	}
	b.n = NewNode(
		b,
		OptNodeChildOf(i),
	)
	return b
}

// BindIncr is the interface a Bind implements
type BindIncr[A any] interface {
	Incr[A]
	Incr() Incr[A]
}

type bindIncr[A, B any] struct {
	n     *Node
	i     Incr[A]
	fn    func(A) Incr[B]
	value Incr[B]
}

func (bi *bindIncr[A, B]) Incr() Incr[B] {
	return bi.value
}

func (bi *bindIncr[A, B]) Stale() bool {
	return true /* we always return true for bind */
}

func (bi *bindIncr[A, B]) Value() B {
	return bi.value.Value()
}

func (bi *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	if err := bi.i.Stabilize(ctx); err != nil {
		return err
	}
	bi.value = bi.fn(bi.i.Value())
	if err := bi.value.Stabilize(ctx); err != nil {
		return err
	}
	return nil

}

func (bi *bindIncr[A, B]) Node() *Node {
	return bi.n
}
