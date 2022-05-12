package incr

import (
	"context"
	"fmt"
)

// Apply2 applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Apply2[A, B, C any](a Incr[A], b Incr[B], fn func(A, B) C) Incr[C] {
	o := &apply2Incr[A, B, C]{
		n: NewNode(),
		a: a,
		b: b,
		fn: func(_ context.Context, a A, b B) (C, error) {
			return fn(a, b), nil
		},
	}
	Link(o, a, b)
	return o
}

// Apply2Context applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Apply2Context[A, B, C any](a Incr[A], b Incr[B], fn func(context.Context, A, B) (C, error)) Incr[C] {
	o := &apply2Incr[A, B, C]{
		n:  NewNode(),
		a:  a,
		b:  b,
		fn: fn,
	}
	Link(o, a, b)
	return o
}

var (
	_ Incr[string] = (*apply2Incr[int, int, string])(nil)
	_ INode        = (*apply2Incr[int, int, string])(nil)
	_ IStabilize   = (*apply2Incr[int, int, string])(nil)
	_ fmt.Stringer = (*apply2Incr[int, int, string])(nil)
)

type apply2Incr[A, B, C any] struct {
	n   *Node
	a   Incr[A]
	b   Incr[B]
	fn  func(context.Context, A, B) (C, error)
	val C
}

func (mn *apply2Incr[A, B, C]) Node() *Node { return mn.n }

func (mn *apply2Incr[A, B, C]) Value() C { return mn.val }

func (mn *apply2Incr[A, B, C]) Stabilize(ctx context.Context) (err error) {
	var val C
	val, err = mn.fn(ctx, mn.a.Value(), mn.b.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *apply2Incr[A, B, C]) String() string {
	return FormatNode(mn.n, "apply2")
}
