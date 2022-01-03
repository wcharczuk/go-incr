package incr

import "context"

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

func (m *cutoffIncr[A]) Value() A {
	return m.value
}

func (m *cutoffIncr[A]) Stabilize(ctx context.Context) error {
	m.initialized = true
	m.value = m.i.Value()
	return nil
}

func (m *cutoffIncr[A]) Stale() bool { return !m.initialized || m.fn(m.value, m.i.Value()) }

func (m *cutoffIncr[A]) getNode() *node {
	return m.n
}
