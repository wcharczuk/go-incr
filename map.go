package incr

import (
	"context"
	"fmt"
)

// Map applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map[A, B any](scope Scope, a Incr[A], fn func(A) B) Incr[B] {
	return MapContext(scope, a, func(_ context.Context, v A) (B, error) {
		return fn(v), nil
	})
}

// MapContext applies a function to a given input incremental and returns
// a new incremental of the output type of that function but is context aware
// and can also return an error, aborting stabilization.
func MapContext[A, B any](scope Scope, a Incr[A], fn func(context.Context, A) (B, error)) Incr[B] {
	return WithinScope(scope, &mapIncr[A, B]{
		n:       NewNode("map"),
		a:       a,
		fn:      fn,
		parents: []INode{a},
	})
}

var (
	_ Incr[string] = (*mapIncr[int, string])(nil)
	_ INode        = (*mapIncr[int, string])(nil)
	_ IStabilize   = (*mapIncr[int, string])(nil)
	_ fmt.Stringer = (*mapIncr[int, string])(nil)
)

type mapIncr[A, B any] struct {
	n       *Node
	a       Incr[A]
	fn      func(context.Context, A) (B, error)
	val     B
	parents []INode
}

func (mn *mapIncr[A, B]) Parents() []INode {
	return mn.parents
}

func (mn *mapIncr[A, B]) Node() *Node {
	return mn.n
}

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
	return mn.n.String()
}
