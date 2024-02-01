package main

import (
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
)

func Custom[T any](ctx context.Context, a incr.Incr[T]) incr.Incr[T] {
	o := &customIncr[T]{
		n: incr.NewNode(),
		a: a,
	}
	incr.Link(o, a)
	return incr.WithBindScope(ctx, o)
}

type customIncr[T any] struct {
	n     *incr.Node
	a     incr.Incr[T]
	value T
}

func (c *customIncr[T]) Value() T         { return c.value }
func (c *customIncr[T]) Node() *incr.Node { return c.n }
func (c *customIncr[T]) Stabilize(ctx context.Context) error {
	c.value = c.a.Value()
	return nil
}

var (
	_ incr.Incr[string] = (*customIncr[string])(nil)
	_ incr.IStabilize   = (*customIncr[string])(nil)
)

func main() {
	ctx := context.Background()
	c := Custom(ctx, incr.Return(ctx, "hello"))
	fmt.Println("before:", c.Value())

	graph := incr.New()
	_ = incr.Observe(graph, c)

	_ = graph.Stabilize(context.Background())
	fmt.Println("after:", c.Value())
}
