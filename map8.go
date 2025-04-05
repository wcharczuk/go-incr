package incr

import (
	"context"
	"fmt"
)

// Map8 applies a function to given input incrementals and returns
// a new incremental of the output type of that function.
func Map8[A, B, C, D, E, F, G, H, I any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], e Incr[E], f Incr[F], g Incr[G], h Incr[H], fn func(A, B, C, D, E, F, G, H) I) Incr[I] {
	return Map8Context(scope, a, b, c, d, e, f, g, h, func(_ context.Context, av A, bv B, cv C, dv D, ev E, fv F, gv G, hv H) (I, error) {
		return fn(av, bv, cv, dv, ev, fv, gv, hv), nil
	})
}

// Map8Context applies a function that accepts a context and returns
// an error, to given input incrementals and returns a
// new incremental of the output type of that function.
func Map8Context[A, B, C, D, E, F, G, H, I any](scope Scope, a Incr[A], b Incr[B], c Incr[C], d Incr[D], e Incr[E], f Incr[F], g Incr[G], h Incr[H], fn func(context.Context, A, B, C, D, E, F, G, H) (I, error)) Incr[I] {
	return WithinScope(scope, &map8Incr[A, B, C, D, E, F, G, H, I]{
		n:       NewNode(KindMap8),
		a:       a,
		b:       b,
		c:       c,
		d:       d,
		e:       e,
		f:       f,
		g:       g,
		h:       h,
		fn:      fn,
		parents: []INode{a, b, c, d, e, f, g, h},
	})
}

var (
	_ Incr[string] = (*map8Incr[int, int, int, int, int, int, int, int, string])(nil)
	_ INode        = (*map8Incr[int, int, int, int, int, int, int, int, string])(nil)
	_ IStabilize   = (*map8Incr[int, int, int, int, int, int, int, int, string])(nil)
	_ fmt.Stringer = (*map8Incr[int, int, int, int, int, int, int, int, string])(nil)
)

type map8Incr[A, B, C, D, E, F, G, H, I any] struct {
	n       *Node
	a       Incr[A]
	b       Incr[B]
	c       Incr[C]
	d       Incr[D]
	e       Incr[E]
	f       Incr[F]
	g       Incr[G]
	h       Incr[H]
	fn      func(context.Context, A, B, C, D, E, F, G, H) (I, error)
	val     I
	parents []INode
}

func (mn *map8Incr[A, B, C, D, E, F, G, H, I]) Parents() []INode {
	return mn.parents
}

func (mn *map8Incr[A, B, C, D, E, F, G, H, I]) Node() *Node { return mn.n }

func (mn *map8Incr[A, B, C, D, E, F, G, H, I]) Value() I { return mn.val }

func (mn *map8Incr[A, B, C, D, E, F, G, H, I]) Stabilize(ctx context.Context) (err error) {
	var val I
	val, err = mn.fn(ctx, mn.a.Value(), mn.b.Value(), mn.c.Value(), mn.d.Value(), mn.e.Value(), mn.f.Value(), mn.g.Value(), mn.h.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *map8Incr[A, B, C, D, E, F, G, H, I]) String() string {
	return mn.n.String()
}
