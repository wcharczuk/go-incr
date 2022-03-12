package incr

import (
	"context"
	"math"
	"testing"
)

func Test_Stabilize_readme(t *testing.T) {
	output := Map2[float64](
		Var(3.14),
		Map(
			Return(10.0),
			func(a float64) float64 {
				return a + 5
			},
		),
		func(a0, a1 float64) float64 {
			return a0 + a1
		},
	)
	_ = Stabilize(
		context.TODO(),
		output,
	)
	itsEqual(t, 18.14, output.Value())
}

func Test_Stabilize_cutoff(t *testing.T) {
	input := Var[float32](3.14)
	output := Map2(
		Cutoff[float32](
			input,
			epsilon[float32](0.1),
		),
		Map(
			Return[float32](10.0),
			addConst[float32](5.0),
		),
		add[float32],
	)

	_ = Stabilize(
		context.TODO(),
		output,
	)
	itsEqual(t, 18.14, output.Value())

	input.Set(3.15)

	_ = Stabilize(
		context.TODO(),
		output,
	)
	itsEqual(t, 18.14, output.Value())

	input.Set(3.26)

	_ = Stabilize(
		context.TODO(),
		output,
	)
	itsEqual(t, 18.26, output.Value())

	_ = Stabilize(
		context.TODO(),
		output,
	)
	itsEqual(t, 18.26, output.Value())
}

func Test_Stabilize_fibonacci(t *testing.T) {
	output := makeFib(32)
	err := Stabilize(
		context.TODO(),
		output,
	)
	itsNil(t, err)
	itsEqual(t, 2178309, output.Value())
}

// epsilon returns a function that returns true
// if the absolute difference of two values is greater
// than a given delta.
func epsilon[A int | float32 | float64](delta A) func(A, A) bool {
	return func(v0, v1 A) bool {
		return math.Abs(float64(v1)-float64(v0)) > float64(delta)
	}
}

// addConst returns a map fn that adds a constant value
// to a given input
func addConst[A int | float32 | float64](v A) func(A) A {
	return func(v0 A) A {
		return v0 + v
	}
}

// add is a map2 fn that adds two values and returns the result
func add[A int | float32 | float64](v0, v1 A) A {
	return v0 + v1
}

func makeFib(height int) (output Incr[int]) {
	prev2 := Return(0) // 0
	prev := Return(1)  // 1
	current := Map2(   // 2
		prev2,
		prev,
		add[int],
	)
	for x := 3; x < height; x++ {
		prev2 = prev
		prev = current
		current = Map2(
			prev2,
			prev,
			add[int],
		)
	}
	output = Map2(
		prev,
		current,
		add[int],
	)
	return
}
