package incr

import (
	"context"
	"fmt"
)

// Map2 applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map2[A, B, C any](ctx context.Context, a Incr[A], b Incr[B], fn func(A, B) C) Incr[C] {
	return Map2Context(ctx, a, b, func(_ context.Context, a A, b B) (C, error) {
		return fn(a, b), nil
	})
}

// Map2Context applies a function that accepts a context and returns an error,
// to a given input incremental and returns a new incremental of
// the output type of that function.
func Map2Context[A, B, C any](ctx context.Context, a Incr[A], b Incr[B], fn func(context.Context, A, B) (C, error)) Incr[C] {
	o := &map2Incr[A, B, C]{
		n:  NewNode(),
		a:  a,
		b:  b,
		fn: fn,
	}
	Link(o, a, b)
	return WithBindScope(ctx, o)
}

var (
	_ Incr[string] = (*map2Incr[int, int, string])(nil)
	_ INode        = (*map2Incr[int, int, string])(nil)
	_ IStabilize   = (*map2Incr[int, int, string])(nil)
	_ fmt.Stringer = (*map2Incr[int, int, string])(nil)
)

type map2Incr[A, B, C any] struct {
	n   *Node
	a   Incr[A]
	b   Incr[B]
	fn  func(context.Context, A, B) (C, error)
	val C
}

func (m2n *map2Incr[A, B, C]) Node() *Node { return m2n.n }

func (m2n *map2Incr[A, B, C]) Value() C { return m2n.val }

func (m2n *map2Incr[A, B, C]) Stabilize(ctx context.Context) (err error) {
	var val C
	val, err = m2n.fn(ctx, m2n.a.Value(), m2n.b.Value())
	if err != nil {
		return
	}
	m2n.val = val
	return nil
}

func (m2n *map2Incr[A, B, C]) String() string {
	return m2n.n.String("map2")
}
