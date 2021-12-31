package incr

import (
	"context"
)

// Map returns a new map incremental.
func Map[A, B any](i Incr[A], fn func(A) B) Incr[B] {
	m := &mapIncr[A, B]{
		i:     i,
		fn:    fn,
		value: fn(i.Value()),
	}
	m.n = newNode(m, optNodeChildOf(i))
	return m
}

// mapIncr is a concrete implementation of Incr for
// the map operator.
type mapIncr[A, B any] struct {
	n     *node
	i     Incr[A]
	value B
	fn    func(A) B
}

// Value implements Incr[B]
func (m *mapIncr[A, B]) Value() B {
	return m.value
}

// Stabilize implements Incr[B]
func (m *mapIncr[A, B]) Stabilize(ctx context.Context) error {
	m.value = m.fn(m.i.Value())
	return nil
}

// getNode implements nodeProvider.
func (m *mapIncr[A, B]) getNode() *node {
	return m.n
}
