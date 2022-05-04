package incr

import "context"

// Map applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map[A, B any](a Incr[A], fn func(a A) (B, error)) Incr[B] {
	n := newNode()
	n.parents = append(n.parents, a)
	output := &mapNode[A, B]{
		n:  n,
		a:  a,
		fn: fn,
	}
	a.Node().children = append(a.Node().children, output)
	return output
}

var (
	_ Incr[string] = (*mapNode[int, string])(nil)
)

type mapNode[A, B any] struct {
	n   *Node
	a   Incr[A]
	fn  func(A) (B, error)
	val B
}

func (mn *mapNode[A, B]) Node() *Node { return mn.n }

func (mn *mapNode[A, B]) Value() B { return mn.val }

func (mn *mapNode[A, B]) Stabilize(ctx context.Context) error {
	nv, err := mn.fn(mn.a.Value())
	if err != nil {
		return err
	}
	mn.val = nv
	return nil
}

func (mn *mapNode[A, B]) String() string {
	return "map[" + mn.n.id.Short() + "]"
}
