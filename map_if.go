package incr

import "context"

func MapIf[A any](a, b Incr[A], p Incr[bool]) Incr[A] {
	o := &mapIfIncr[A]{
		n: NewNode(),
		a: a,
		b: b,
		p: p,
	}
	Link(o, a, b, p)
	return o
}

var (
	_ Incr[string] = (*mapIfIncr[string])(nil)
	_ GraphNode    = (*mapIfIncr[string])(nil)
	_ Stabilizer   = (*mapIfIncr[string])(nil)
)

type mapIfIncr[A any] struct {
	n     *Node
	a     Incr[A]
	b     Incr[A]
	p     Incr[bool]
	value A
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

func (mi *mapIfIncr[A]) String() string { return "map_if[" + mi.n.id.Short() + "]" }
