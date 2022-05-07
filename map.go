package incr

import "context"

// Map applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map[A, B any](a Incr[A], fn func(a A) B) Incr[B] {
	n := NewNode()
	m := &mapNode[A, B]{
		n:  n,
		a:  a,
		fn: fn,
	}
	n.AddChildren(a)
	a.Node().AddParents(m)
	return m
}

var (
	_ Incr[string] = (*mapNode[int, string])(nil)
	_ GraphNode    = (*mapNode[int, string])(nil)
	_ Stabilizer   = (*mapNode[int, string])(nil)
)

type mapNode[A, B any] struct {
	n   *Node
	a   Incr[A]
	fn  func(A) B
	val B
}

func (mn *mapNode[A, B]) Node() *Node { return mn.n }

func (mn *mapNode[A, B]) Value() B { return mn.val }

func (mn *mapNode[A, B]) Stabilize(ctx context.Context) error {
	mn.val = mn.fn(mn.a.Value())
	return nil
}

func (mn *mapNode[A, B]) String() string {
	return "map[" + mn.n.id.Short() + "]"
}
