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
	m.n = newNode(
		m,
		optNodeChildOf(i0),
		optNodeChildOf(i1),
		optNodeChildOf(i2),
	)
	return m
}

type map3Incr[A, B, C, D comparable] struct {
	n     *node
	i0    Incr[A]
	i1    Incr[B]
	i2    Incr[C]
	fn    func(A, B, C) D
	value D
}

func (m *map3Incr[A, B, C, D]) Value() D {
	return m.value
}

func (m *map3Incr[A, B, C, D]) Stabilize(ctx context.Context) error {
	m.value = m.fn(m.i0.Value(), m.i1.Value(), m.i2.Value())
	return nil
}

func (m *map3Incr[A, B, C, D]) getValue() any {
	return m.Value()
}

func (m *map3Incr[A, B, C, D]) getNode() *node {
	return m.n
}
