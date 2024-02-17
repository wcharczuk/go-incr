package incr

import (
	"context"
	"fmt"
)

// Freeze yields an incremental that takes the value of an
// input incremental and doesn't change thereafter.
func Freeze[A any](scope Scope, i Incr[A]) Incr[A] {
	return WithinScope(scope, &freezeIncr[A]{
		n: NewNode("freeze"),
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
}

func (f *freezeIncr[T]) Parents() []INode {
	return []INode{f.i}
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
