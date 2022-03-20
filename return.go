package incr

import "context"

// Return creates a new incr from a given value.
//
// You can think of this as a constant.
func Return[A any](value A) Incr[A] {
	r := &returnIncr[A]{
		value: value,
	}
	r.n = NewNode(r)
	return r
}

type returnIncr[A any] struct {
	n     *Node
	value A
}

func (r *returnIncr[A]) Value() A {
	return r.value
}

func (r *returnIncr[A]) Stale() bool {
	return false
}

func (r *returnIncr[A]) Stabilize(_ context.Context) (bool, error) {
	return false, nil
}

func (r *returnIncr[A]) Node() *Node { return r.n }
