package main

import (
	"context"
	"fmt"

	incr "github.com/wcharczuk/go-incremental"
)

func main() {
	output := incr.Map2[float64](
		incr.Var(3.14),
		incr.Map(
			incr.Return(10.0),
			func(a float64) float64 {
				return a + 5
			},
		),
		func(a0, a1 float64) float64 {
			return a0 + a1
		},
	)
	_ = incr.Stabilize(
		incr.WithTracing(context.Background()),
		output,
	)
	fmt.Println(output.Value()) // prints 18.14
}
