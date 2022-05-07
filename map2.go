package incr

import "context"

// Map2 applies a function to a given input incremental and returns
// a new incremental of the output type of that function.
func Map2[A, B, C any](a Incr[A], b Incr[B], fn func(A, B) C) Incr[C] {
	o := &map2Node[A, B, C]{
		n:  NewNode(),
		a:  a,
		b:  b,
		fn: fn,
	}
	Link(o, a, b)
	return o
}

var (
	_ Incr[string] = (*map2Node[int, int, string])(nil)
	_ GraphNode    = (*map2Node[int, int, string])(nil)
	_ Stabilizer   = (*map2Node[int, int, string])(nil)
)

type map2Node[A, B, C any] struct {
	n   *Node
	a   Incr[A]
	b   Incr[B]
	fn  func(A, B) C
	val C
}

func (mn *map2Node[A, B, C]) Node() *Node { return mn.n }

func (mn *map2Node[A, B, C]) Value() C { return mn.val }

func (mn *map2Node[A, B, C]) Stabilize(ctx context.Context) error {
	mn.val = mn.fn(mn.a.Value(), mn.b.Value())
	return nil
}

func (mn *map2Node[A, B, C]) String() string {
	return "map2[" + mn.n.id.Short() + "]"
}
