package main

import (
	"context"
	"fmt"
	"os"

	incr "github.com/wcharczuk/go-incremental"
)

func main() {
	input := incr.Var("this")

	output :=
		incr.Map2(
			incr.Map[string](
				input,
				func(a string) string {
					return "not " + a
				},
			),
			incr.Return("test"),
			func(a, b string) string {
				return a + " " + b
			},
		)
	if err := incr.Stabilize(
		incr.WithTracing(context.Background()),
		output,
	); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
	fmt.Printf("value: %s\n", output.Value())
}
