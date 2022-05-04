package incr

import "context"

// Return yields an incremental for a given value.
func Return[T any](v T) Incr[T] {
	return &returnIncr[T]{
		n: newNode(),
		v: v,
	}
}

type returnIncr[T any] struct {
	n *Node
	v T
}

func (r returnIncr[T]) Node() *Node { return r.n }

func (r returnIncr[T]) Value() T { return r.v }

func (r returnIncr[T]) Stabilize(_ context.Context) error { return nil }

func (r returnIncr[T]) String() string { return "return[" + r.n.id.Short() + "]" }
