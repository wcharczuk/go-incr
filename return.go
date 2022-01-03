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
	n           *node
	initialized bool
	value       A
}

func (r *returnIncr[A]) Value() A {
	return r.value
}

func (r *returnIncr[A]) Stabilize(_ context.Context) error {
	r.initialized = true
	return nil
}

func (r *returnIncr[A]) Stale() bool { return !r.initialized }

func (r *returnIncr[A]) getNode() *node {
	return r.n
}
