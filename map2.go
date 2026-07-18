package incr

import (
	"context"
	"fmt"
)

// Map2 applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map2[A, B, C any](scope Scope, a Incr[A], b Incr[B], fn func(A, B) C) Incr[C] {
	// Its own type rather than adapting fn to the context signature, which would allocate a
	// closure per node purely to wrap it; see [Map] for the reasoning.
	m := &map2PlainIncr[A, B, C]{
		n:  scope.newNode(KindMap2),
		a:  a,
		b:  b,
		fn: fn,
	}
	m.parents[0] = a
	m.parents[1] = b
	return WithinScope(scope, m)
}

var (
	_ Incr[string] = (*map2PlainIncr[int, int, string])(nil)
	_ INode        = (*map2PlainIncr[int, int, string])(nil)
	_ IStabilize   = (*map2PlainIncr[int, int, string])(nil)
	_ IParents     = (*map2PlainIncr[int, int, string])(nil)
	_ fmt.Stringer = (*map2PlainIncr[int, int, string])(nil)
)

type map2PlainIncr[A, B, C any] struct {
	n   *Node
	a   Incr[A]
	b   Incr[B]
	fn  func(A, B) C
	val C

	// parents is an array rather than a slice so that constructing the node does
	// not allocate a separate input list; [Parents] hands out a slice over it.
	parents [2]INode
}

func (m2n *map2PlainIncr[A, B, C]) Parents() []INode { return m2n.parents[:] }

func (m2n *map2PlainIncr[A, B, C]) Node() *Node { return m2n.n }

func (m2n *map2PlainIncr[A, B, C]) Value() C { return m2n.val }

func (m2n *map2PlainIncr[A, B, C]) Stabilize(_ context.Context) error {
	m2n.val = m2n.fn(m2n.a.Value(), m2n.b.Value())
	return nil
}

func (m2n *map2PlainIncr[A, B, C]) String() string { return m2n.n.String() }

// Map2Context applies a function that accepts a context and returns an error,
// to a given input incremental and returns a new incremental of
// the output type of that function.
func Map2Context[A, B, C any](scope Scope, a Incr[A], b Incr[B], fn func(context.Context, A, B) (C, error)) Incr[C] {
	m := &map2Incr[A, B, C]{
		n:  scope.newNode(KindMap2),
		a:  a,
		b:  b,
		fn: fn,
	}
	m.parents[0] = a
	m.parents[1] = b
	return WithinScope(scope, m)
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

	// parents is an array rather than a slice so that constructing the node does
	// not allocate a separate input list; [Parents] hands out a slice over it.
	parents [2]INode
}

func (m2n *map2Incr[A, B, C]) Parents() []INode {
	return m2n.parents[:]
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
	return m2n.n.String()
}
