package incr

import "context"

func MapIf[A any](a, b Incr[A], p Incr[bool]) Incr[A] {
	n := newNode()
	mi := &mapIfIncr[A]{
		n: n,
		a: a,
		b: b,
		p: p,
	}
	n.children = append(n.children, a, b, p)
	a.Node().parents = append(a.Node().parents, mi)
	b.Node().parents = append(b.Node().parents, mi)
	p.Node().parents = append(p.Node().parents, mi)
	return mi
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
