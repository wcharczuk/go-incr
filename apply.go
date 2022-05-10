package incr

import (
	"context"
	"fmt"
)

// Apply applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Apply[A, B any](a Incr[A], fn func(context.Context, A) (B, error)) Incr[B] {
	m := &applyIncr[A, B]{
		n:  NewNode(),
		a:  a,
		fn: fn,
	}
	Link(m, a)
	return m
}

var (
	_ Incr[string] = (*applyIncr[int, string])(nil)
	_ INode        = (*applyIncr[int, string])(nil)
	_ IStabilize   = (*applyIncr[int, string])(nil)
	_ fmt.Stringer = (*applyIncr[int, string])(nil)
)

type applyIncr[A, B any] struct {
	n   *Node
	a   Incr[A]
	fn  func(context.Context, A) (B, error)
	val B
}

func (mn *applyIncr[A, B]) Node() *Node { return mn.n }

func (mn *applyIncr[A, B]) Value() B { return mn.val }

func (mn *applyIncr[A, B]) Stabilize(ctx context.Context) (err error) {
	var val B
	val, err = mn.fn(ctx, mn.a.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *applyIncr[A, B]) String() string {
	return FormatNode(mn.n, "apply")
}
