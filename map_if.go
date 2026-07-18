package incr

import (
	"context"
	"fmt"
)

// MapIf returns an incremental that yields one of two values
// based on the boolean condition returned from a third incremental.
//
// Specifically, we term this [MapIf] because the nodes are all
// linked in the graph, but the value changes during stabilization.
func MapIf[A any](scope Scope, a, b Incr[A], p Incr[bool]) Incr[A] {
	return WithinScope(scope, &mapIfIncr[A]{
		n: scope.newNode(KindMapIf),
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
	// parents is the storage [Parents] fills and returns a slice over, so that
	// asking a node for its inputs does not allocate a fresh list every call.
	parents [3]INode
}

func (mi *mapIfIncr[T]) Parents() []INode {
	mi.parents[0] = mi.a
	mi.parents[1] = mi.b
	mi.parents[2] = mi.p
	return mi.parents[:]
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
