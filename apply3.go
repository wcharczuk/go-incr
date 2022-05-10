package incr

import (
	"context"
	"fmt"
)

// Apply3 applies a function to given input incrementals and returns
// a new incremental of the output type of that function.
func Apply3[A, B, C, D any](a Incr[A], b Incr[B], c Incr[C], fn func(context.Context, A, B, C) (D, error)) Incr[D] {
	o := &apply3Incr[A, B, C, D]{
		n:  NewNode(),
		a:  a,
		b:  b,
		c:  c,
		fn: fn,
	}
	Link(o, a, b, c)
	return o
}

var (
	_ Incr[string] = (*apply3Incr[int, int, int, string])(nil)
	_ INode        = (*apply3Incr[int, int, int, string])(nil)
	_ IStabilize   = (*apply3Incr[int, int, int, string])(nil)
	_ fmt.Stringer = (*apply3Incr[int, int, int, string])(nil)
)

type apply3Incr[A, B, C, D any] struct {
	n   *Node
	a   Incr[A]
	b   Incr[B]
	c   Incr[C]
	fn  func(context.Context, A, B, C) (D, error)
	val D
}

func (mn *apply3Incr[A, B, C, D]) Node() *Node { return mn.n }

func (mn *apply3Incr[A, B, C, D]) Value() D { return mn.val }

func (mn *apply3Incr[A, B, C, D]) Stabilize(ctx context.Context) (err error) {
	var val D
	val, err = mn.fn(ctx, mn.a.Value(), mn.b.Value(), mn.c.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *apply3Incr[A, B, C, D]) String() string {
	return FormatNode(mn.n, "apply3")
}
