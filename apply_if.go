package incr

import (
	"context"
	"fmt"
)

// ApplyIf returns an incremental that yields one of two values
// based on the boolean condition returned from a third.
//
// Specifically, we term this _Apply_If because the nodes are all
// linked in the graph, but the value changes during stabilization.
func ApplyIf[A any](a, b Incr[A], p Incr[bool]) Incr[A] {
	o := &applyIfIncr[A]{
		n: NewNode(),
		a: a,
		b: b,
		p: p,
	}
	Link(o, a, b, p)
	return o
}

var (
	_ Incr[string] = (*applyIfIncr[string])(nil)
	_ INode        = (*applyIfIncr[string])(nil)
	_ IStabilize   = (*applyIfIncr[string])(nil)
	_ fmt.Stringer = (*applyIfIncr[string])(nil)
)

type applyIfIncr[A any] struct {
	n     *Node
	a     Incr[A]
	b     Incr[A]
	p     Incr[bool]
	value A
}

func (mi *applyIfIncr[A]) Node() *Node { return mi.n }

func (mi *applyIfIncr[A]) Value() A {
	return mi.value
}

func (mi *applyIfIncr[A]) Stabilize(ctx context.Context) error {
	if mi.p.Value() {
		mi.value = mi.a.Value()
	} else {
		mi.value = mi.b.Value()
	}
	return nil
}

func (mi *applyIfIncr[A]) String() string { return FormatNode(mi.n, "apply_if") }
