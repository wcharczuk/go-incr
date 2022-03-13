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
	// we set the changed at here so that stabilization
	// passes correctly pick up that the variable has changed.
	v.n.changedAt = v.n.recomputedAt + 1
	v.latest = value
}

func (v *varIncr[A]) Value() A {
	return v.value
}

func (v *varIncr[A]) Stabilize(ctx context.Context, _ Generation) error {
	v.value = v.latest
	return nil
}

func (v *varIncr[A]) Node() *Node { return v.n }
