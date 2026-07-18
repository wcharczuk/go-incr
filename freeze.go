package incr

import (
	"context"
	"fmt"
)

// Freeze yields an incremental that takes the value of an
// input incremental on the first stabilization and
// doesn't change thereafter.
//
// Stabilization propagates through this node even
// after the first stabilization.
func Freeze[A any](scope Scope, i Incr[A]) Incr[A] {
	return WithinScope(scope, &freezeIncr[A]{
		n: scope.newNode(KindFreeze),
		i: i,
	})
}

var (
	_ Incr[string] = (*freezeIncr[string])(nil)
	_ IStabilize   = (*freezeIncr[string])(nil)
	_ INode        = (*freezeIncr[string])(nil)
	_ fmt.Stringer = (*freezeIncr[string])(nil)
)

type freezeIncr[A any] struct {
	n        *Node
	i        Incr[A]
	freezeAt uint64
	v        A
	// parents is the storage [Parents] fills and returns a slice over, so that
	// asking a node for its inputs does not allocate a fresh list every call.
	parents [1]INode
}

func (f *freezeIncr[T]) Parents() []INode {
	f.parents[0] = f.i
	return f.parents[:]
}

func (f *freezeIncr[T]) Node() *Node { return f.n }

func (f *freezeIncr[T]) Value() T { return f.v }

func (f *freezeIncr[T]) String() string { return f.n.String() }

func (f *freezeIncr[A]) Stabilize(_ context.Context) error {
	if f.freezeAt > 0 {
		return nil
	}
	f.freezeAt = GraphForNode(f).stabilizationNum
	f.v = f.i.Value()
	return nil
}
