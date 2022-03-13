package incr

import (
	"context"
)

// Var returns a new variable that wraps a given value.
func Var[A any](value A) VarIncr[A] {
	v := &varIncr[A]{
		latest: value,
	}
	v.n = NewNode(v)
	return v
}

// VarIncr extends incr with a Watch method.
type VarIncr[A any] interface {
	Incr[A]
	Watch() WatchIncr[A]
	Set(A)
}

type varIncr[A any] struct {
	n      *Node
	value  A
	latest A
	stale  bool
}

func (v *varIncr[A]) Watch() WatchIncr[A] {
	return Watch[A](v)
}

func (v *varIncr[A]) Set(value A) {
	v.latest = value
	v.stale = true
}

func (v *varIncr[A]) Stale() bool {
	return v.stale
}

func (v *varIncr[A]) Value() A {
	return v.value
}

func (v *varIncr[A]) Stabilize(ctx context.Context) error {
	v.value = v.latest
	v.stale = false
	return nil
}

func (v *varIncr[A]) Node() *Node { return v.n }
