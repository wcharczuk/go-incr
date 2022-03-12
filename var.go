package incr

import (
	"context"
)

// Var returns a new variable that wraps a given value.
func Var[A comparable](value A) VarIncr[A] {
	v := &varIncr[A]{
		latest: value,
	}
	v.n = newNode(v)
	v.n.changedAt = generation(1)
	return v
}

// VarIncr extends incr with a Watch method.
type VarIncr[A comparable] interface {
	Incr[A]
	Watch() WatchIncr[A]
	Set(A)
}

type varIncr[A comparable] struct {
	n      *node
	value  A
	latest A
}

func (v *varIncr[A]) Watch() WatchIncr[A] {
	return Watch[A](v)
}

func (v *varIncr[A]) Set(value A) {
	v.n.changedAt = v.n.recomputedAt + 1
	v.latest = value
}

func (v *varIncr[A]) Value() A {
	return v.value
}

func (v *varIncr[A]) Stabilize(ctx context.Context) error {
	v.value = v.latest
	return nil
}

func (v *varIncr[A]) getValue() any { return v.Value() }

func (v *varIncr[A]) getNode() *node { return v.n }
