package incr

import (
	"context"
)

// Cutoff returns a new wrapping cutoff incremental.
//
// The goal of the cutoff incremental is to stop recomputation at a given
// node if the difference between the previous and latest values are not
// significant enough to warrant a full recomputation of the children of this node.
func Cutoff[A comparable](i Incr[A], fn func(value, latest A) bool) Incr[A] {
	n := newNode()
	co := &cutoffIncr[A]{
		i:  i,
		fn: fn,
	}
	n.cutoff = co.Cutoff
	n.children = append(n.children, i)
	i.Node().parents = append(i.Node().parents, co)
	return co
}

var (
	_ Incr[string] = (*cutoffIncr[string])(nil)
	_ GraphNode    = (*cutoffIncr[string])(nil)
	_ Stabilizer   = (*cutoffIncr[string])(nil)
	_ Cutoffer     = (*cutoffIncr[string])(nil)
)

// cutoffIncr is a concrete implementation of Incr for
// the cutoff operator.
type cutoffIncr[A comparable] struct {
	n     *Node
	i     Incr[A]
	value A
	fn    func(A, A) bool
}

func (c *cutoffIncr[A]) Value() A {
	return c.value
}

func (c *cutoffIncr[A]) Stabilize(ctx context.Context) error {
	c.value = c.i.Value()
	return nil
}

func (c *cutoffIncr[A]) Cutoff(ctx context.Context) bool {
	return c.fn(c.value, c.i.Value())
}

func (c *cutoffIncr[A]) Node() *Node {
	return c.n
}

func (c *cutoffIncr[A]) String() string { return "cutoff[" + c.n.id.Short() + "]" }
