package main

import (
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
)

func Custom[T any](scope incr.Scope, a incr.Incr[T]) incr.Incr[T] {
	return incr.WithinScope(scope, &customIncr[T]{
		n: incr.NewNode("custom"),
		a: a,
	})
}

type customIncr[T any] struct {
	n     *incr.Node
	a     incr.Incr[T]
	value T
}

func (c *customIncr[T]) Parents() []incr.INode {
	return []incr.INode{c.a}
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
	_ incr.IParents     = (*customIncr[string])(nil)
)

func main() {
	ctx := context.Background()
	graph := incr.New()
	c := Custom(graph, incr.Return(graph, "hello"))
	fmt.Println("before:", c.Value())

	_ = incr.MustObserve(graph, c)

	_ = graph.Stabilize(ctx)
	fmt.Println("after:", c.Value())
}
