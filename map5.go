package incr

import (
	"context"
	"fmt"
)

// Map5 applies a function to given input incrementals and returns
// a new incremental of the output type of that function.
func Map5[A, B, C, D, E, F any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], e Incr[E], fn func(A, B, C, D, E) F) Incr[F] {
	return Map5Context(scope, a, b, c, d, e, func(_ context.Context, av A, bv B, cv C, dv D, ev E) (F, error) {
		return fn(av, bv, cv, dv, ev), nil
	})
}

// Map5Context applies a function that accepts a context and returns
// an error, to given input incrementals and returns a
// new incremental of the output type of that function.
func Map5Context[A, B, C, D, E, F any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], e Incr[E], fn func(context.Context, A, B, C, D, E) (F, error)) Incr[F] {
	return WithinScope(scope, &map5Incr[A, B, C, D, E, F]{
		n:       NewNode(KindMap5),
		a:       a,
		b:       b,
		c:       c,
		d:       d,
		e:       e,
		fn:      fn,
		parents: []INode{a, b, c, d, e},
	})
}

var (
	_ Incr[string] = (*map5Incr[int, int, int, int, int, string])(nil)
	_ INode        = (*map5Incr[int, int, int, int, int, string])(nil)
	_ IStabilize   = (*map5Incr[int, int, int, int, int, string])(nil)
	_ fmt.Stringer = (*map5Incr[int, int, int, int, int, string])(nil)
)

type map5Incr[A, B, C, D, E, F any] struct {
	n       *Node
	a       Incr[A]
	b       Incr[B]
	c       Incr[C]
	d       Incr[D]
	e       Incr[E]
	fn      func(context.Context, A, B, C, D, E) (F, error)
	val     F
	parents []INode
}

func (mn *map5Incr[A, B, C, D, E, F]) Parents() []INode {
	return mn.parents
}

func (mn *map5Incr[A, B, C, D, E, F]) Node() *Node { return mn.n }

func (mn *map5Incr[A, B, C, D, E, F]) Value() F { return mn.val }

func (mn *map5Incr[A, B, C, D, E, F]) Stabilize(ctx context.Context) (err error) {
	var val F
	val, err = mn.fn(ctx, mn.a.Value(), mn.b.Value(), mn.c.Value(), mn.d.Value(), mn.e.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *map5Incr[A, B, C, D, E, F]) String() string {
	return mn.n.String()
}
