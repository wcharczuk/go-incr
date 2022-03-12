package main

import (
	"context"
	"fmt"
	"os"

	incr "github.com/wcharczuk/go-incremental"
)

func add(v0, v1 int) int { return v0 + v1 }

func makeFib(n int) incr.Incr[int] {
	switch n {
	case 0:
		return incr.Return(0)
	case 1:
		return incr.Return(1)
	}
	cache := make([]incr.Incr[int], n+1)
	cache[0] = incr.Return(0) // 0
	cache[1] = incr.Return(1) // 1
	if n > 1 {
		cache[2] = incr.Return(1) // 1
	}
	if n > 2 {
		cache[3] = incr.Return(2) // 2
	}
	if n > 3 {
		cache[4] = incr.Return(3) // 3
	}
	return _makeFib(n, cache)
}

func _makeFib(n int, cache []incr.Incr[int]) (output incr.Incr[int]) {
	if cache[n] != nil {
		return cache[n]
	}
	var prev2 incr.Incr[int]
	if cache[n-2] != nil {
		prev2 = cache[n-2]
	} else {
		prev2 = _makeFib(n-2, cache)
	}
	var prev incr.Incr[int]
	if cache[n-1] != nil {
		prev = cache[n-1]
	} else {
		prev = _makeFib(n-1, cache)
	}
	cache[n] = incr.Map2(
		prev2,
		prev,
		add,
	)
	return cache[n]
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
