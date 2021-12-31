package incr

import (
	"context"
)

// MapIf returns one value or the other as the result of a given boolean incremental.
func MapIf[A any](i0, i1 Incr[A], c Incr[bool]) Incr[A] {
	mi := &mapIfIncr[A]{
		i0: i0,
		i1: i1,
		c:  c,
	}
	mi.n = newNode(
		mi,
		optNodeChildOf(i0),
		optNodeChildOf(i1),
	)
	return mi
}

type mapIfIncr[A any] struct {
	n     *node
	i0    Incr[A]
	i1    Incr[A]
	c     Incr[bool]
	value A
}

// Value implements Incr[A]
func (mii mapIfIncr[A]) Value() A {
	return mii.value
}

// Stabilize implements Incr[A]
func (mii mapIfIncr[A]) Stabilize(ctx context.Context) error {
	if mii.c.Value() {
		mii.value = mii.i0.Value()
	} else {
		mii.value = mii.i1.Value()
	}
	return nil
}

func (mii mapIfIncr[A]) getNode() *node {
	return mii.n
}
