package incr

import (
	"context"
	"fmt"
)

// Map2 applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map2[A, B, C any](a Incr[A], b Incr[B], fn func(context.Context, A, B) (C, error)) Incr[C] {
	o := &map2Incr[A, B, C]{
		n:  NewNode(),
		a:  a,
		b:  b,
		fn: fn,
	}
	Link(o, a, b)
	return o
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

func (mn *map2Incr[A, B, C]) Node() *Node { return mn.n }

func (mn *map2Incr[A, B, C]) Value() C { return mn.val }

func (mn *map2Incr[A, B, C]) Stabilize(ctx context.Context) (err error) {
	var val C
	val, err = mn.fn(ctx, mn.a.Value(), mn.b.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *map2Incr[A, B, C]) String() string {
	return FormatNode(mn.n, "map2")
}
