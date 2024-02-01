package incr

import (
	"context"
	"fmt"
)

// Cutoff2 returns a new cutoff incremental that takes an epsilon input.
func Cutoff2[A, B any](ctx context.Context, epsilon Incr[A], input Incr[B], fn Cutoff2Func[A, B]) Incr[B] {
	return Cutoff2Context[A, B](ctx, epsilon, input, func(_ context.Context, epsilon A, oldv, newv B) (bool, error) {
		return fn(epsilon, oldv, newv), nil
	})
}

// Cutoff2Context returns a new cutoff incremental that takes an epsilon input.
//
// The goal of the cutoff incremental is to stop recomputation at a given
// node if the difference between the previous and latest values are not
// significant enough to warrant a full recomputation of the children of this node.
func Cutoff2Context[A, B any](ctx context.Context, epsilon Incr[A], input Incr[B], fn Cutoff2ContextFunc[A, B]) Cutoff2Incr[A, B] {
	o := &cutoff2Incr[A, B]{
		n:  NewNode(),
		fn: fn,
		e:  epsilon,
		i:  input,
	}
	// we short circuit setup of the node cutoff reference here.
	// this can be discovered in initialization but saves a step.
	Link(o, input)
	Link(o, epsilon)
	return WithBindScope(ctx, o)
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

func (c *cutoff2Incr[A, B]) String() string { return c.n.String("cutoff2") }
