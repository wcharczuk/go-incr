package incr

import (
	"context"
	"fmt"
	"testing"
)

func Benchmark_Stabilize_withPreInitialize(b *testing.B) {
	// create a 2048 node, 12 level reverse tree of functions
	size := 2048
	nodes := make([]Incr[string], size)
	for x := 0; x < size; x++ {
		nodes[x] = Var(fmt.Sprintf("var_%d", x))
	}

	var cursor int
	for x := size; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Map2(nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	gs := nodes[len(nodes)-1]
	Initialize(testContext(), gs)
	benchStabilize(gs, b)
}

func concat(_ context.Context, a, b string) (string, error) {
	return a + b, nil
}

func benchStabilize(gs INode, b *testing.B) {
	for n := 0; n < b.N; n++ {
		benchStabilizeSingle(gs, n)
	}
}

func benchStabilizeSingle(gs INode, n int) {
	_ = Stabilize(testContext(), gs)
}
