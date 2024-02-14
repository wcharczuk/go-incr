package incr

import (
	"context"
	"fmt"
)

// Map4 applies a function to given input incrementals and returns
// a new incremental of the output type of that function.
func Map4[A, B, C, D, E any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], fn func(A, B, C, D) E) Incr[E] {
	return Map4Context(scope, a, b, c, d, func(_ context.Context, av A, bv B, cv C, dv D) (E, error) {
		return fn(av, bv, cv, dv), nil
	})
}

// Map4Context applies a function that accepts a context and returns
// an error, to given input incrementals and returns a
// new incremental of the output type of that function.
func Map4Context[A, B, C, D, E any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], fn func(context.Context, A, B, C, D) (E, error)) Incr[E] {
	return WithinScope(scope, &map4Incr[A, B, C, D, E]{
		n:       NewNode("map4"),
		a:       a,
		b:       b,
		c:       c,
		d:       d,
		fn:      fn,
		parents: []INode{a, b, c, d},
	})
}

var (
	_ Incr[string] = (*map4Incr[int, int, int, int, string])(nil)
	_ INode        = (*map4Incr[int, int, int, int, string])(nil)
	_ IStabilize   = (*map4Incr[int, int, int, int, string])(nil)
	_ fmt.Stringer = (*map4Incr[int, int, int, int, string])(nil)
)

type map4Incr[A, B, C, D, E any] struct {
	n       *Node
	a       Incr[A]
	b       Incr[B]
	c       Incr[C]
	d       Incr[D]
	fn      func(context.Context, A, B, C, D) (E, error)
	val     E
	parents []INode
}

func (mn *map4Incr[A, B, C, D, E]) Parents() []INode {
	return mn.parents
}

func (mn *map4Incr[A, B, C, D, E]) Node() *Node { return mn.n }

func (mn *map4Incr[A, B, C, D, E]) Value() E { return mn.val }

func (mn *map4Incr[A, B, C, D, E]) Stabilize(ctx context.Context) (err error) {
	var val E
	val, err = mn.fn(ctx, mn.a.Value(), mn.b.Value(), mn.c.Value(), mn.d.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *map4Incr[A, B, C, D, E]) String() string {
	return mn.n.String()
}
