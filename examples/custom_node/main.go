package main

import (
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
)

func Custom[T any](a incr.Incr[T]) incr.Incr[T] {
	o := &customIncr[T]{
		n: incr.NewNode(),
		a: a,
	}
	incr.Link(o, a)
	return o
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
	_ incr.Stabilizer   = (*customIncr[string])(nil)
)

func main() {
	c := Custom(incr.Return("hello"))
	fmt.Println("before:", c.Value())
	_ = incr.Stabilize(context.Background(), c)
	fmt.Println("after:", c.Value())
}
