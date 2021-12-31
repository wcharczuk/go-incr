package incr

import "context"

// Return creates a new incr from a given value.
//
// You can think of this as a constant.
func Return[A any](value A) Incr[A] {
	r := &returnIncr[A]{
		value: value,
	}
	r.n = newNode(r)
	return r
}

type returnIncr[A any] struct {
	n     *node
	value A
}

// Value implements Incr[A].
func (r *returnIncr[A]) Value() A {
	return r.value
}

// Stabilize implements Incr[A].
func (r *returnIncr[A]) Stabilize(_ context.Context) error { return nil }

// getNode implements node provider.
func (r *returnIncr[A]) getNode() *node {
	return r.n
}
