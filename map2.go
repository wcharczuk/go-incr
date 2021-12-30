package incr

import "context"

// Map returns a new map incremental.
func Map2[A, B, C any](i0 Incr[A], i1 Incr[B], fn func(A, B) C) Incr[C] {
	m2 := &map2Incr[A, B, C]{
		i0:    i0,
		i1:    i1,
		fn:    fn,
		value: fn(i0.Value(), i1.Value()),
	}
	m2.node = newNode(
		m2,
		optNodeChildOf(i0),
		optNodeChildOf(i1),
	)
	return m2
}

type map2Incr[A, B, C any] struct {
	*node
	i0    Incr[A]
	i1    Incr[B]
	fn    func(A, B) C
	value C
}

func (m *map2Incr[A, B, C]) Value() C {
	return m.value
}

func (m *map2Incr[A, B, C]) Stabilize(ctx context.Context) error {
	m.value = m.fn(m.i0.Value(), m.i1.Value())
	return nil
}

func (m *map2Incr[A, B, C]) getNode() *node {
	return m.node
}
