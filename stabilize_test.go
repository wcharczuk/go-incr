package incr

import (
	"context"
	"fmt"
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

func Test_Stabilize_error(t *testing.T) {
	var shouldError bool
	output := Map2(
		Func(func(_ context.Context) (v float64, err error) {
			if shouldError {
				err = fmt.Errorf("this is just a test")
				return
			}
			v = 3.14
			return
		}),
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
	err := Stabilize(
		context.TODO(),
		output,
	)
	itsNil(t, err)
	itsEqual(t, 18.14, output.Value())

	shouldError = true
}
