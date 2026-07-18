package incr

import (
	"context"
	"fmt"
)

// Map applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map[A, B any](scope Scope, a Incr[A], fn func(A) B) Incr[B] {
	// A distinct type rather than adapting fn to the context signature. The adapter is a
	// closure capturing fn, which is a third allocation per node on top of the node struct
	// and its metadata, and it buys nothing: this is the most common combinator in most
	// graphs. A field holding both function types and a branch in Stabilize would also
	// work, but that trades an allocation paid once at construction for a test paid on
	// every recompute for the life of the node.
	m := &mapPlainIncr[A, B]{
		n:  scope.newNode(KindMap),
		a:  a,
		fn: fn,
	}
	m.parents[0] = a
	return WithinScope(scope, m)
}

var (
	_ Incr[string] = (*mapPlainIncr[int, string])(nil)
	_ INode        = (*mapPlainIncr[int, string])(nil)
	_ IStabilize   = (*mapPlainIncr[int, string])(nil)
	_ IParents     = (*mapPlainIncr[int, string])(nil)
	_ fmt.Stringer = (*mapPlainIncr[int, string])(nil)
)

type mapPlainIncr[A, B any] struct {
	n   *Node
	a   Incr[A]
	fn  func(A) B
	val B
	// parents is stored as an array rather than a slice so that constructing the
	// node does not allocate a separate one-element input list; [Parents] hands
	// out a slice over it.
	parents [1]INode
}

func (mn *mapPlainIncr[A, B]) Parents() []INode { return mn.parents[:] }

func (mn *mapPlainIncr[A, B]) Node() *Node { return mn.n }

func (mn *mapPlainIncr[A, B]) Value() B { return mn.val }

func (mn *mapPlainIncr[A, B]) Stabilize(_ context.Context) error {
	mn.val = mn.fn(mn.a.Value())
	return nil
}

func (mn *mapPlainIncr[A, B]) String() string { return mn.n.String() }

// MapContext applies a function to a given input incremental and returns
// a new incremental of the output type of that function but is context aware
// and can also return an error, aborting stabilization.
func MapContext[A, B any](scope Scope, a Incr[A], fn func(context.Context, A) (B, error)) Incr[B] {
	m := &mapIncr[A, B]{
		n:  scope.newNode(KindMap),
		a:  a,
		fn: fn,
	}
	m.parents[0] = a
	return WithinScope(scope, m)
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
	// parents is stored as an array rather than a slice so that constructing the
	// node does not allocate a separate one-element input list; [Parents] hands
	// out a slice over it.
	parents [1]INode
}

func (mn *mapIncr[A, B]) Parents() []INode {
	return mn.parents[:]
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
