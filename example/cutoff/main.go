package main

import (
	"context"
	"fmt"
	"math"

	incr "github.com/wcharczuk/go-incremental"
)

func epsilon(delta float64) func(float64, float64) bool {
	return func(v0, v1 float64) bool {
		return math.Abs(v1-v0) > delta
	}
}

func addConst(v float64) func(float64) float64 {
	return func(v0 float64) float64 {
		return v0 + v
	}
}

func add(v0, v1 float64) float64 {
	return v0 + v1
}

func main() {
	input := incr.Var(3.14)
	output := incr.Map2(
		incr.Cutoff[float64](
			input,
			epsilon(0.1),
		),
		incr.Map(
			incr.Return(10.0),
			addConst(5),
		),
		add,
	)

	_ = incr.Stabilize(
		context.Background(),
		output,
	)
	fmt.Printf("output: %0.2f\n", output.Value())

	fmt.Println("input.Set(3.15)")
	input.Set(3.15)

	_ = incr.Stabilize(
		context.Background(),
		output,
	)
	fmt.Printf("output: %0.2f\n", output.Value())

	fmt.Println("input.Set(3.26)")
	input.Set(3.26)

	_ = incr.Stabilize(
		incr.WithTracing(context.Background()),
		output,
	)
	fmt.Printf("output: %0.2f\n", output.Value())

	_ = incr.Stabilize(
		context.Background(),
		output,
	)
	fmt.Printf("output: %0.2f\n", output.Value())
}
