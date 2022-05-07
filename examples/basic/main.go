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

	output := incr.Map2[string, string](v0, v1, func(a, b string) string { return a + " and " + b })

	if err := incr.Stabilize(context.Background(), output); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
