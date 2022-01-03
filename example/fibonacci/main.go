package main

import (
	"context"
	"fmt"
	"os"

	incr "github.com/wcharczuk/go-incremental"
)

func add(v0, v1 int) int { return v0 + v1 }

func makeFib(height int) (output incr.Incr[int]) {
	prev2 := incr.Return(0) // 0
	prev := incr.Return(1)  // 1
	current := incr.Map2(   // 2
		prev2,
		prev,
		add,
	)
	for x := 3; x < height; x++ {
		prev2 = prev
		prev = current
		current = incr.Map2(
			prev2,
			prev,
			add,
		)
	}
	output = incr.Map2(
		prev,
		current,
		add,
	)
	return
}

func main() {
	fmt.Println("creating computation")
	output := makeFib(10)
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
