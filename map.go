package incr

import (
	"context"
)

// Map returns a new map incremental.
//
// Map applies a given function `fn` to a given input incremental.
//
// Map holds the resulting value of the computation for re-use.
func Map[A, B comparable](i Incr[A], fn func(A) B) Incr[B] {
	m := &mapIncr[A, B]{
		i:  i,
		fn: fn,
	}
	m.n = NewNode(m, OptNodeChildOf(i))
	return m
}

// mapIncr is a concrete implementation of Incr for
// the map operator.
type mapIncr[A, B comparable] struct {
	n     *Node
	i     Incr[A]
	fn    func(A) B
	value B
}

func (m *mapIncr[A, B]) Value() B {
	return m.value
}

func (m *mapIncr[A, B]) Stabilize(ctx context.Context, g Generation) error {
	oldValue := m.value
	m.value = m.fn(m.i.Value())
	if oldValue != m.value {
		m.n.changedAt = g
	}
	return nil
}

func (m *mapIncr[A, B]) Node() *Node {
	return m.n
}
