package incr

import (
	"context"
)

// Var returns a new variable that wraps a given value.
func Var[A comparable](value A) VarIncr[A] {
	v := &varIncr[A]{
		latest: value,
	}
	v.n = NewNode(v)
	return v
}

// VarIncr extends incr with a Watch method.
type VarIncr[A comparable] interface {
	Incr[A]
	Watch() WatchIncr[A]
	Set(A)
}

type varIncr[A comparable] struct {
	n      *Node
	value  A
	latest A
}

func (v *varIncr[A]) Watch() WatchIncr[A] {
	return Watch[A](v)
}

func (v *varIncr[A]) Set(value A) {
	v.latest = value
}

func (v *varIncr[A]) Value() A {
	return v.value
}

func (v *varIncr[A]) Stabilize(ctx context.Context, g Generation) error {
	if v.value != v.latest {
		v.n.changedAt = g
	}
	v.value = v.latest
	return nil
}

func (v *varIncr[A]) Node() *Node { return v.n }
