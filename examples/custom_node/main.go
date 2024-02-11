package main

import (
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
)

func Custom[T any](scope incr.Scope, a incr.Incr[T]) incr.Incr[T] {
	o := &customIncr[T]{
		n: incr.NewNode("custom"),
		a: a,
	}
	incr.Link(o, a)
	return incr.WithinScope(scope, o)
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
	graph := incr.New()
	c := Custom(graph, incr.Return(graph, "hello"))
	fmt.Println("before:", c.Value())

	_ = incr.Observe(graph, c)

	_ = graph.Stabilize(ctx)
	fmt.Println("after:", c.Value())
}
