package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wcharczuk/go-incr"
)

func main() {
	v0 := incr.Var(incr.Root(), "foo")
	v1 := incr.Var(incr.Root(), "bar")
	output := incr.Map2(incr.Root(), v0, v1, func(a, b string) string { return a + " and " + b })

	graph := incr.New()
	observer := incr.Observe(incr.Root(), graph, output)

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
