package incr

import "context"

// Return creates a new incr from a given value.
//
// You can think of this as a constant.
func Return[A comparable](value A) Incr[A] {
	r := &returnIncr[A]{
		value: value,
	}
	r.n = newNode(r)
	r.n.changedAt = generation(1)
	return r
}

type returnIncr[A comparable] struct {
	n     *node
	value A
}

func (r *returnIncr[A]) Value() A {
	return r.value
}

func (r *returnIncr[A]) Stabilize(_ context.Context) error {
	return nil
}

func (r *returnIncr[A]) getValue() any { return r.Value() }

func (r *returnIncr[A]) getNode() *node { return r.n }
