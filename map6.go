package incr

import (
	"context"
	"fmt"
)

// Map6 applies a function to given input incrementals and returns
// a new incremental of the output type of that function.
func Map6[A, B, C, D, E, F, G any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], e Incr[E], f Incr[F], fn func(A, B, C, D, E, F) G) Incr[G] {
	return Map6Context(scope, a, b, c, d, e, f, func(_ context.Context, av A, bv B, cv C, dv D, ev E, fv F) (G, error) {
		return fn(av, bv, cv, dv, ev, fv), nil
	})
}

// Map6Context applies a function that accepts a context and returns
// an error, to given input incrementals and returns a
// new incremental of the output type of that function.
func Map6Context[A, B, C, D, E, F, G any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], e Incr[E], f Incr[F], fn func(context.Context, A, B, C, D, E, F) (G, error)) Incr[G] {
	return WithinScope(scope, &map6Incr[A, B, C, D, E, F, G]{
		n:       NewNode("map6"),
		a:       a,
		b:       b,
		c:       c,
		d:       d,
		e:       e,
		f:       f,
		fn:      fn,
		parents: []INode{a, b, c, d, e, f},
	})
}

var (
	_ Incr[string] = (*map6Incr[int, int, int, int, int, int, string])(nil)
	_ INode        = (*map6Incr[int, int, int, int, int, int, string])(nil)
	_ IStabilize   = (*map6Incr[int, int, int, int, int, int, string])(nil)
	_ fmt.Stringer = (*map6Incr[int, int, int, int, int, int, string])(nil)
)

type map6Incr[A, B, C, D, E, F, G any] struct {
	n       *Node
	a       Incr[A]
	b       Incr[B]
	c       Incr[C]
	d       Incr[D]
	e       Incr[E]
	f       Incr[F]
	fn      func(context.Context, A, B, C, D, E, F) (G, error)
	val     G
	parents []INode
}

func (mn *map6Incr[A, B, C, D, E, F, G]) Parents() []INode {
	return mn.parents
}

func (mn *map6Incr[A, B, C, D, E, F, G]) Node() *Node { return mn.n }

func (mn *map6Incr[A, B, C, D, E, F, G]) Value() G { return mn.val }

func (mn *map6Incr[A, B, C, D, E, F, G]) Stabilize(ctx context.Context) (err error) {
	var val G
	val, err = mn.fn(ctx, mn.a.Value(), mn.b.Value(), mn.c.Value(), mn.d.Value(), mn.e.Value(), mn.f.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *map6Incr[A, B, C, D, E, F, G]) String() string {
	return mn.n.String()
}
