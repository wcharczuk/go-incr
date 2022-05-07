package incr

import "context"

// Map3 applies a function to given input incrementals and returns
// a new incremental of the output type of that function.
func Map3[A, B, C, D any](a Incr[A], b Incr[B], c Incr[C], fn func(context.Context, A, B, C) (D, error)) Incr[D] {
	o := &map3Incr[A, B, C, D]{
		n:  NewNode(),
		a:  a,
		b:  b,
		c:  c,
		fn: fn,
	}
	Link(o, a, b, c)
	return o
}

var (
	_ Incr[string] = (*map3Incr[int, int, int, string])(nil)
	_ GraphNode    = (*map3Incr[int, int, int, string])(nil)
	_ Stabilizer   = (*map3Incr[int, int, int, string])(nil)
)

type map3Incr[A, B, C, D any] struct {
	n   *Node
	a   Incr[A]
	b   Incr[B]
	c   Incr[C]
	fn  func(context.Context, A, B, C) (D, error)
	val D
}

func (mn *map3Incr[A, B, C, D]) Node() *Node { return mn.n }

func (mn *map3Incr[A, B, C, D]) Value() D { return mn.val }

func (mn *map3Incr[A, B, C, D]) Stabilize(ctx context.Context) (err error) {
	var val D
	val, err = mn.fn(ctx, mn.a.Value(), mn.b.Value(), mn.c.Value())
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *map3Incr[A, B, C, D]) String() string {
	return "map3[" + mn.n.id.Short() + "]"
}
