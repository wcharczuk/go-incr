package incr

import (
	"context"
)

// MapIf returns one value or the other as the result of a given boolean incremental.
func MapIf[A comparable](i0, i1 Incr[A], c Incr[bool]) Incr[A] {
	m := &mapIfIncr[A]{
		i0: i0,
		i1: i1,
		c:  c,
	}
	m.n = newNode(
		m,
		optNodeChildOf(i0),
		optNodeChildOf(i1),
		optNodeChildOf(c),
	)
	return m
}

type mapIfIncr[A comparable] struct {
	n     *node
	i0    Incr[A]
	i1    Incr[A]
	c     Incr[bool]
	value A
}

func (m *mapIfIncr[A]) Value() A {
	return m.value
}

func (m *mapIfIncr[A]) Stabilize(ctx context.Context) error {
	if m.c.Value() {
		m.value = m.i0.Value()
	} else {
		m.value = m.i1.Value()
	}
	return nil
}

func (m *mapIfIncr[A]) getValue() any {
	return m.Value()
}

func (m *mapIfIncr[A]) getNode() *node {
	return m.n
}
