package incr

import (
	"context"
	"time"
)

// Var returns a new variable that wraps a given value.
func Var[A any](value A) VarIncr[A] {
	v := new(varIncr[A])
	v.n = newNode(v)
	v.value = value
	return v
}

// VarIncr extends incr with a Watch method.
type VarIncr[A any] interface {
	Incr[A]
	Watch() WatchIncr[A]
	Set(A)
}

type varIncr[A any] struct {
	n      *node
	value  A
	latest A
}

func (v *varIncr[A]) Watch() WatchIncr[A] {
	return Watch[A](v)
}

func (v *varIncr[A]) Set(value A) {
	v.latest = value
	v.n.changedAt = time.Now()
}

func (v *varIncr[A]) Value() A {
	return v.value
}

func (v *varIncr[A]) Stabilize(ctx context.Context) error {
	var zero A
	v.value = v.latest
	v.latest = zero
	return nil
}

func (v *varIncr[A]) getNode() *node { return v.n }
