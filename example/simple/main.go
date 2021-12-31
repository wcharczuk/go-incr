package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"

	incr "github.com/wcharczuk/go-incremental"
)

func main() {
	var input float64 = prompt("input")

	output := incr.Map2[float64](
		incr.Var(input),
		incr.MapIf(
			incr.Return(2.0),
			incr.Return(4.0),
			incr.Func(func(_ context.Context) (bool, error) {
				println("input", input)
				return input > 5, nil
			}),
		),
		func(v0, v1 float64) float64 {
			return v0 * v1
		},
	)
	if err := incr.Stabilize(
		incr.WithTracing(context.Background()),
		output,
	); err != nil {
		fatal(err)
	}
	fmt.Printf("value: %0.2f\n", output.Value())
}

func prompt(message string) (output float64) {
	fmt.Printf("%s: ", message)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		rawValue := scanner.Text()
		output, _ = strconv.ParseFloat(rawValue, 64)
		return
	}
	return
}

func fatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
