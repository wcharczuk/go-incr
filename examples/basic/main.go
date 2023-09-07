package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wcharczuk/go-incr"
)

func main() {
	v0 := incr.Var("foo")
	v1 := incr.Var("bar")
	output := incr.Map2(v0, v1, func(a, b string) string { return a + " and " + b })

	graph := incr.New()
	observer := incr.MustObserve(graph, output)

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
