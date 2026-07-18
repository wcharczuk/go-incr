package incr

import (
	"context"
	"fmt"
)

// Map7 applies a function to given input incrementals and returns
// a new incremental of the output type of that function.
func Map7[A, B, C, D, E, F, G, H any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], e Incr[E], f Incr[F], g Incr[G], fn func(A, B, C, D, E, F, G) H) Incr[H] {
	return Map7Context(scope, a, b, c, d, e, f, g, func(_ context.Context, av A, bv B, cv C, dv D, ev E, fv F, gv G) (H, error) {
		return fn(av, bv, cv, dv, ev, fv, gv), nil
	})
}

// Map7Context applies a function that accepts a context and returns
// an error, to given input incrementals and returns a
// new incremental of the output type of that function.
func Map7Context[A, B, C, D, E, F, G, H any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], e Incr[E], f Incr[F], g Incr[G], fn func(context.Context, A, B, C, D, E, F, G) (H, error)) Incr[H] {
	m := &map7Incr[A, B, C, D, E, F, G, H]{
		n:  scope.newNode(KindMap7),
		a:  a,
		b:  b,
		c:  c,
		d:  d,
		e:  e,
		f:  f,
		g:  g,
		fn: fn,
	}
	m.parents[0] = a
	m.parents[1] = b
	m.parents[2] = c
	m.parents[3] = d
	m.parents[4] = e
	m.parents[5] = f
	m.parents[6] = g
	return WithinScope(scope, m)
}

var (
	_ Incr[string] = (*map7Incr[int, int, int, int, int, int, int, string])(nil)
	_ INode        = (*map7Incr[int, int, int, int, int, int, int, string])(nil)
	_ IStabilize   = (*map7Incr[int, int, int, int, int, int, int, string])(nil)
	_ fmt.Stringer = (*map7Incr[int, int, int, int, int, int, int, string])(nil)
)

type map7Incr[A, B, C, D, E, F, G, H any] struct {
	n   *Node
	a   Incr[A]
	b   Incr[B]
	c   Incr[C]
	d   Incr[D]
	e   Incr[E]
	f   Incr[F]
	g   Incr[G]
	fn  func(context.Context, A, B, C, D, E, F, G) (H, error)
	val H
	// parents is an array rather than a slice so that constructing the node does
	// not allocate a separate input list; [Parents] hands out a slice over it.
	parents [7]INode
}

func (mn *map7Incr[A, B, C, D, E, F, G, H]) Parents() []INode {
	return mn.parents[:]
}

func (mn *map7Incr[A, B, C, D, E, F, G, H]) Node() *Node { return mn.n }

func (mn *map7Incr[A, B, C, D, E, F, G, H]) Value() H { return mn.val }

func (mn *map7Incr[A, B, C, D, E, F, G, H]) Stabilize(ctx context.Context) (err error) {
	var val H
	val, err = mn.fn(ctx, mn.a.Value(), mn.b.Value(), mn.c.Value(), mn.d.Value(), mn.e.Value(), mn.f.Value(), mn.g.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *map7Incr[A, B, C, D, E, F, G, H]) String() string {
	return mn.n.String()
}
