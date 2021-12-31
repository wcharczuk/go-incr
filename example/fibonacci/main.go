package main

import (
	"context"
	"fmt"
	"os"

	incr "github.com/wcharczuk/go-incremental"
)

func makeFib(height int) incr.Incr[int] {
	if height == 0 {
		return incr.Return(0)
	}
	if height == 1 {
		return incr.Return(1)
	}
	return incr.Map2(
		makeFib(height-2),
		makeFib(height-1),
		func(v0, v1 int) int {
			return v0 + v1
		},
	)
}

func main() {
	fmt.Println("creating computation")
	output := makeFib(32)
	fmt.Println("stabilizing computation")
	if err := incr.Stabilize(
		incr.WithTracing(context.Background()),
		output,
	); err != nil {
		fatal(err)
	}
	fmt.Printf("value: %d\n", output.Value())
}

func fatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}
