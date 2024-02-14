package incr

import (
	"context"
	"fmt"
)

// Cutoff2 returns a new cutoff incremental that takes an epsilon input.
func Cutoff2[A, B any](bs Scope, epsilon Incr[A], input Incr[B], fn Cutoff2Func[A, B]) Incr[B] {
	return Cutoff2Context[A, B](bs, epsilon, input, func(_ context.Context, epsilon A, oldv, newv B) (bool, error) {
		return fn(epsilon, oldv, newv), nil
	})
}

// Cutoff2Context returns a new cutoff incremental that takes an epsilon input.
//
// The goal of the cutoff incremental is to stop recomputation at a given
// node if the difference between the previous and latest values are not
// significant enough to warrant a full recomputation of the children of this node.
func Cutoff2Context[A, B any](bs Scope, epsilon Incr[A], input Incr[B], fn Cutoff2ContextFunc[A, B]) Cutoff2Incr[A, B] {
	return WithinScope(bs, &cutoff2Incr[A, B]{
		n:  NewNode("cutoff2"),
		fn: fn,
		e:  epsilon,
		i:  input,
	})
}

// CutoffIncr is an incremental node that implements the ICutoff interface.
type Cutoff2Incr[A, B any] interface {
	Incr[B]
	IStabilize
	ICutoff
}

var (
	_ Incr[string]             = (*cutoff2Incr[int, string])(nil)
	_ Cutoff2Incr[int, string] = (*cutoff2Incr[int, string])(nil)
	_ IStabilize               = (*cutoff2Incr[int, string])(nil)
	_ ICutoff                  = (*cutoff2Incr[int, string])(nil)
	_ fmt.Stringer             = (*cutoff2Incr[int, string])(nil)
)

// Cutoff2Func is a function that implements cutoff checking.
type Cutoff2Func[A, B any] func(A, B, B) bool

// Cutoff2ContextFunc is a function that implements cutoff checking
// and takes a context.
type Cutoff2ContextFunc[A, B any] func(context.Context, A, B, B) (bool, error)

// cutoffIncr is a concrete implementation of Incr for
// the cutoff operator.
type cutoff2Incr[A, B any] struct {
	n     *Node
	e     Incr[A]
	i     Incr[B]
	value B
	fn    Cutoff2ContextFunc[A, B]
}

func (c *cutoff2Incr[A, B]) Parents() []INode {
	return []INode{c.e, c.i}
}

func (c *cutoff2Incr[A, B]) Value() B {
	return c.value
}

func (c *cutoff2Incr[A, B]) Stabilize(ctx context.Context) error {
	c.value = c.i.Value()
	return nil
}

func (c *cutoff2Incr[A, B]) Cutoff(ctx context.Context) (bool, error) {
	return c.fn(ctx, c.e.Value(), c.value, c.i.Value())
}

func (c *cutoff2Incr[A, B]) Node() *Node {
	return c.n
}

func (c *cutoff2Incr[A, B]) String() string { return c.n.String() }
