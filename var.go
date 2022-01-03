package incr

import (
	"context"
)

// Var returns a new variable that wraps a given value.
func Var[A comparable](value A) VarIncr[A] {
	v := new(varIncr[A])
	v.n = newNode(v)
	v.latest = value
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
	v.latest = value
}

func (v *varIncr[A]) Value() A {
	return v.value
}

func (v *varIncr[A]) Stabilize(ctx context.Context) error {
	v.value = v.latest
	return nil
}

func (v *varIncr[A]) Stale() bool {
	return v.latest != v.value
}

func (v *varIncr[A]) getNode() *node { return v.n }
