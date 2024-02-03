package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wcharczuk/go-incr"
)

func main() {
	ctx := context.Background()
	v0 := incr.Var(ctx, "foo")
	v1 := incr.Var(ctx, "bar")
	output := incr.Map2(ctx, v0, v1, func(a, b string) string { return a + " and " + b })

	graph := incr.New()
	observer := incr.Observe(ctx, graph, output)

	if err := graph.Stabilize(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	fmt.Println("output:", observer.Value())

	v0.Set("not-foo")

	if err := graph.Stabilize(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	fmt.Println("output:", observer.Value())
}
