package incr

import (
	"context"
	"fmt"
)

// Cutoff returns a new wrapping cutoff incremental.
//
// The goal of the cutoff incremental is to stop recomputation at a given
// node if the difference between the previous and latest values are not
// significant enough to warrant a full recomputation of the children of this node.
func Cutoff[A any](i Incr[A], fn func(A, A) bool) Incr[A] {
	o := &cutoffIncr[A]{
		n: NewNode(),
		i: i,
		fn: func(_ context.Context, oldv, newv A) bool {
			return fn(oldv, newv)
		},
	}
	// we short circuit setup of the node cutoff reference here.
	// this can be discovered in initialization but saves a step.
	o.Node().cutoff = o.Cutoff
	Link(o, i)
	return o
}

// CutoffContext returns a new wrapping cutoff incremental.
//
// The goal of the cutoff incremental is to stop recomputation at a given
// node if the difference between the previous and latest values are not
// significant enough to warrant a full recomputation of the children of this node.
func CutoffContext[A any](i Incr[A], fn func(context.Context, A, A) bool) Incr[A] {
	o := &cutoffIncr[A]{
		n:  NewNode(),
		i:  i,
		fn: fn,
	}
	// we short circuit setup of the node cutoff reference here.
	// this can be discovered in initialization but saves a step.
	o.Node().cutoff = o.Cutoff
	Link(o, i)
	return o
}

var (
	_ Incr[string] = (*cutoffIncr[string])(nil)
	_ INode        = (*cutoffIncr[string])(nil)
	_ IStabilize   = (*cutoffIncr[string])(nil)
	_ ICutoff      = (*cutoffIncr[string])(nil)
	_ fmt.Stringer = (*cutoffIncr[string])(nil)
)

// cutoffIncr is a concrete implementation of Incr for
// the cutoff operator.
type cutoffIncr[A any] struct {
	n     *Node
	i     Incr[A]
	value A
	fn    func(context.Context, A, A) bool
}

func (c *cutoffIncr[A]) Value() A {
	return c.value
}

func (c *cutoffIncr[A]) Stabilize(ctx context.Context) error {
	c.value = c.i.Value()
	return nil
}

func (c *cutoffIncr[A]) Cutoff(ctx context.Context) bool {
	return c.fn(ctx, c.value, c.i.Value())
}

func (c *cutoffIncr[A]) Node() *Node {
	return c.n
}

func (c *cutoffIncr[A]) String() string { return c.n.String("cutoff") }
