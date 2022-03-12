package incr

import "context"

// Map3 returns a new map3 incremental.
func Map3[A, B, C, D comparable](i0 Incr[A], i1 Incr[B], i2 Incr[C], fn func(A, B, C) D) Incr[D] {
	m := &map3Incr[A, B, C, D]{
		i0: i0,
		i1: i1,
		i2: i2,
		fn: fn,
	}
	m.n = NewNode(
		m,
		OptNodeChildOf(i0),
		OptNodeChildOf(i1),
		OptNodeChildOf(i2),
	)
	return m
}

type map3Incr[A, B, C, D comparable] struct {
	n     *Node
	i0    Incr[A]
	i1    Incr[B]
	i2    Incr[C]
	fn    func(A, B, C) D
	value D
}

func (m *map3Incr[A, B, C, D]) Value() D {
	return m.value
}

func (m *map3Incr[A, B, C, D]) Stabilize(ctx context.Context, g Generation) error {
	oldValue := m.value
	m.value = m.fn(m.i0.Value(), m.i1.Value(), m.i2.Value())
	if oldValue != m.value {
		m.n.changedAt = g
	}
	return nil
}

func (m *map3Incr[A, B, C, D]) Node() *Node {
	return m.n
}
