package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wcharczuk/go-incr"
)

func main() {
	g := incr.New()
	v0 := incr.Var(g, "foo")
	v1 := incr.Var(g, "bar")
	output := incr.Map2(g, v0, v1, func(a, b string) string { return a + " and " + b })

	observer := incr.MustObserve(g, output)

	if err := g.Stabilize(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	fmt.Println("output:", observer.Value())

	v0.Set("not-foo")

	if err := g.Stabilize(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	fmt.Println("output:", observer.Value())
}
