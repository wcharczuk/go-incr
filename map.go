package incr

import (
	"context"
	"fmt"
)

// Map applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map[A, B any](a Incr[A], fn func(context.Context, A) (B, error)) Incr[B] {
	m := &mapIncr[A, B]{
		n:  NewNode(),
		a:  a,
		fn: fn,
	}
	Link(m, a)
	return m
}

var (
	_ Incr[string] = (*mapIncr[int, string])(nil)
	_ INode        = (*mapIncr[int, string])(nil)
	_ IStabilize   = (*mapIncr[int, string])(nil)
	_ fmt.Stringer = (*mapIncr[int, string])(nil)
)

type mapIncr[A, B any] struct {
	n   *Node
	a   Incr[A]
	fn  func(context.Context, A) (B, error)
	val B
}

func (mn *mapIncr[A, B]) Node() *Node { return mn.n }

func (mn *mapIncr[A, B]) Value() B { return mn.val }

func (mn *mapIncr[A, B]) Stabilize(ctx context.Context) (err error) {
	var val B
	val, err = mn.fn(ctx, mn.a.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *mapIncr[A, B]) String() string {
	return FormatNode(mn.n, "map")
}
