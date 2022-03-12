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
	m.n = NewNode(
		m,
		OptNodeChildOf(i0),
		OptNodeChildOf(i1),
		OptNodeChildOf(c),
	)
	return m
}

type mapIfIncr[A comparable] struct {
	n     *Node
	i0    Incr[A]
	i1    Incr[A]
	c     Incr[bool]
	value A
}

func (m *mapIfIncr[A]) Value() A {
	return m.value
}

func (m *mapIfIncr[A]) Stabilize(ctx context.Context, g Generation) error {
	oldValue := m.value
	if m.c.Value() {
		m.value = m.i0.Value()
	} else {
		m.value = m.i1.Value()
	}
	if oldValue != m.value {
		m.n.changedAt = g
	}
	return nil
}

func (m *mapIfIncr[A]) Node() *Node {
	return m.n
}
