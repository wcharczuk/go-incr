package incr

import (
	"context"
)

// Cutoff returns a new wrapping cutoff incremental.
//
// The goal of the cutoff incremental is to stop recomputation at a given
// node if the difference between the current and updated values are not
// significant.
func Cutoff[A any](i Incr[A], fn func(value A, latest A) bool) Incr[A] {
	co := &cutoffIncr[A]{
		i:  i,
		fn: fn,
	}
	co.n = newNode(co, optNodeChildOf(i))
	return co
}

// cutoffIncr is a concrete implementation of Incr for
// the cutoff operator.
type cutoffIncr[A any] struct {
	n           *node
	i           Incr[A]
	fn          func(A, A) bool
	initialized bool
	value       A
}

func (c *cutoffIncr[A]) Value() A {
	return c.value
}

func (c *cutoffIncr[A]) Stabilize(ctx context.Context) error {
	c.initialized = true
	c.value = c.i.Value()
	return nil
}

func (c *cutoffIncr[A]) Stale() (stale bool) {
	stale = !c.initialized || c.fn(c.value, c.i.Value())
	return
}

func (c *cutoffIncr[A]) getNode() *node {
	return c.n
}
