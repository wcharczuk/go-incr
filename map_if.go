package incr

import (
	"context"
	"fmt"
)

// MapIf returns an incremental that yields one of two values
// based on the boolean condition returned from a third.
//
// Specifically, we term this _Apply_If because the nodes are all
// linked in the graph, but the value changes during stabilization.
func MapIf[A any](scope Scope, a, b Incr[A], p Incr[bool]) Incr[A] {
	return WithinScope(scope, &mapIfIncr[A]{
		n: NewNode("map_if"),
		a: a,
		b: b,
		p: p,
	})
}

var (
	_ Incr[string] = (*mapIfIncr[string])(nil)
	_ INode        = (*mapIfIncr[string])(nil)
	_ IStabilize   = (*mapIfIncr[string])(nil)
	_ fmt.Stringer = (*mapIfIncr[string])(nil)
)

type mapIfIncr[A any] struct {
	n     *Node
	a     Incr[A]
	b     Incr[A]
	p     Incr[bool]
	value A
}

func (mi *mapIfIncr[T]) Parents() []INode {
	return []INode{mi.a, mi.b, mi.p}
}

func (mi *mapIfIncr[A]) Node() *Node { return mi.n }

func (mi *mapIfIncr[A]) Value() A {
	return mi.value
}

func (mi *mapIfIncr[A]) Stabilize(ctx context.Context) error {
	if mi.p.Value() {
		mi.value = mi.a.Value()
	} else {
		mi.value = mi.b.Value()
	}
	return nil
}

func (mi *mapIfIncr[A]) String() string { return mi.n.String() }
