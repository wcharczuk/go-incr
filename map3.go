package incr

import "context"

// Map3 applies a function to given input incrementals and returns
// a new incremental of the output type of that function.
func Map3[A, B, C, D any](a Incr[A], b Incr[B], c Incr[C], fn func(A, B, C) (D, error)) Incr[D] {
	n := newNode()
	o := &map3Node[A, B, C, D]{
		n:  n,
		a:  a,
		b:  b,
		c:  c,
		fn: fn,
	}
	n.children = append(n.children, a, b, c)
	a.Node().parents = append(a.Node().parents, o)
	b.Node().parents = append(b.Node().parents, o)
	c.Node().parents = append(c.Node().parents, o)
	return o
}

var (
	_ Incr[string] = (*map3Node[int, int, int, string])(nil)
	_ GraphNode    = (*map3Node[int, int, int, string])(nil)
	_ Stabilizer   = (*map3Node[int, int, int, string])(nil)
)

type map3Node[A, B, C, D any] struct {
	n   *Node
	a   Incr[A]
	b   Incr[B]
	c   Incr[C]
	fn  func(A, B, C) (D, error)
	val D
}

func (mn *map3Node[A, B, C, D]) Node() *Node { return mn.n }

func (mn *map3Node[A, B, C, D]) Value() D { return mn.val }

func (mn *map3Node[A, B, C, D]) Stabilize(ctx context.Context) error {
	nv, err := mn.fn(mn.a.Value(), mn.b.Value(), mn.c.Value())
	if err != nil {
		return err
	}
	mn.val = nv
	return nil
}

func (mn *map3Node[A, B, C, D]) String() string {
	return "map3[" + mn.n.id.Short() + "]"
}
