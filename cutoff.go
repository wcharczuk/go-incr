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
func Cutoff[A any](bs Scope, i Incr[A], fn CutoffFunc[A]) Incr[A] {
	return CutoffContext[A](bs, i, func(_ context.Context, oldv, newv A) (bool, error) {
		return fn(oldv, newv), nil
	})
}

// CutoffContext returns a new wrapping cutoff incremental.
//
// The goal of the cutoff incremental is to stop recomputation at a given
// node if the difference between the previous and latest values are not
// significant enough to warrant a full recomputation of the children of this node.
func CutoffContext[A any](bs Scope, i Incr[A], fn CutoffContextFunc[A]) Incr[A] {
	return WithinScope(bs, &cutoffIncr[A]{
		n:  NewNode("cutoff"),
		i:  i,
		fn: fn,
	})
}

// CutoffFunc is a function that implements cutoff checking.
type CutoffFunc[A any] func(A, A) bool

// CutoffContextFunc is a function that implements cutoff checking
// and takes a context.
type CutoffContextFunc[A any] func(context.Context, A, A) (bool, error)

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
	fn    CutoffContextFunc[A]
}

func (c *cutoffIncr[A]) Parents() []INode {
	return []INode{c.i}
}

func (c *cutoffIncr[A]) Value() A {
	return c.value
}

func (c *cutoffIncr[A]) Stabilize(ctx context.Context) error {
	c.value = c.i.Value()
	return nil
}

func (c *cutoffIncr[A]) Cutoff(ctx context.Context) (bool, error) {
	return c.fn(ctx, c.value, c.i.Value())
}

func (c *cutoffIncr[A]) Node() *Node {
	return c.n
}

func (c *cutoffIncr[A]) String() string { return c.n.String() }
