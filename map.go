package incr

import "context"

// Map applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map[A, B any](a Incr[A], fn func(a A) B) Incr[B] {
	m := &mapIncr[A, B]{
		n:  NewNode(),
		a:  a,
		fn: fn,
	}
	Link(m, a)
	return m
}

var (
	_ Incr[string] = (*mapIncr[int, string])(nil)
	_ GraphNode    = (*mapIncr[int, string])(nil)
	_ Stabilizer   = (*mapIncr[int, string])(nil)
)

type mapIncr[A, B any] struct {
	n   *Node
	a   Incr[A]
	fn  func(A) B
	val B
}

func (mn *mapIncr[A, B]) Node() *Node { return mn.n }

func (mn *mapIncr[A, B]) Value() B { return mn.val }

func (mn *mapIncr[A, B]) Stabilize(ctx context.Context) error {
	mn.val = mn.fn(mn.a.Value())
	return nil
}

func (mn *mapIncr[A, B]) String() string {
	return "map[" + mn.n.id.Short() + "]"
}
