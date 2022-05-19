package incr

import (
	"context"
	"fmt"
)

// Freeze yields an incremental that takes the value of an
// input incremental and doesn't change thereafter.
func Freeze[A any](i Incr[A]) Incr[A] {
	o := &freezeIncr[A]{
		n: NewNode(),
		i: i,
	}
	Link(o, i)
	return o
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

func (f *freezeIncr[T]) Node() *Node { return f.n }

func (f *freezeIncr[T]) Value() T { return f.v }

func (f *freezeIncr[T]) String() string { return Label(f.n, "freeze") }

func (f *freezeIncr[A]) Stabilize(_ context.Context) error {
	if f.freezeAt > 0 {
		return nil
	}
	f.freezeAt = f.n.g.stabilizationNum
	f.v = f.i.Value()
	return nil
}
