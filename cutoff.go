package incr

import (
	"context"
)

// Cutoff returns a new wrapping cutoff incremental.
//
// The goal of the cutoff incremental is to stop recomputation at a given
// node if the difference between the previous and latest values are not
// significant enough to warrant a full recomputation of the children of this node.
func Cutoff[A comparable](i Incr[A], fn func(value A, latest A) bool) Incr[A] {
	co := &cutoffIncr[A]{
		i:  i,
		fn: fn,
	}
	co.n = NewNode(co, OptNodeChildOf(i))
	return co
}

// cutoffIncr is a concrete implementation of Incr for
// the cutoff operator.
type cutoffIncr[A comparable] struct {
	n     *Node
	i     Incr[A]
	fn    func(A, A) bool
	value A
}

func (c *cutoffIncr[A]) Value() A {
	return c.value
}

func (c *cutoffIncr[A]) Stale() bool { return false }

func (c *cutoffIncr[A]) Stabilize(ctx context.Context) error {
	newValue := c.i.Value()
	if c.fn(c.value, newValue) {
		c.value = c.i.Value()
	}
	return nil
}

func (c *cutoffIncr[A]) Node() *Node {
	return c.n
}
