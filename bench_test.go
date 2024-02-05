package incr

import (
	"fmt"
	"math/rand"
	"testing"
)

func Benchmark_Stabilize_withPreInitialize_512(b *testing.B) {
	benchmarkSize(512, b)
}

func Benchmark_Stabilize_withPreInitialize_1024(b *testing.B) {
	benchmarkSize(1024, b)
}

func Benchmark_Stabilize_withPreInitialize_2048(b *testing.B) {
	benchmarkSize(2048, b)
}

func Benchmark_Stabilize_withPreInitialize_4096(b *testing.B) {
	benchmarkSize(4096, b)
}

func Benchmark_Stabilize_withPreInitialize_8192(b *testing.B) {
	benchmarkSize(8192, b)
}

func Benchmark_Stabilize_withPreInitialize_16384(b *testing.B) {
	benchmarkSize(16384, b)
}

func Benchmark_ParallelStabilize_withPreInitialize_512(b *testing.B) {
	benchmarkParallelSize(512, b)
}

func Benchmark_ParallelStabilize_withPreInitialize_1024(b *testing.B) {
	benchmarkParallelSize(1024, b)
}

func Benchmark_ParallelStabilize_withPreInitialize_2048(b *testing.B) {
	benchmarkParallelSize(2048, b)
}

func Benchmark_ParallelStabilize_withPreInitialize_4096(b *testing.B) {
	benchmarkParallelSize(4096, b)
}

func Benchmark_ParallelStabilize_withPreInitialize_8192(b *testing.B) {
	benchmarkParallelSize(8192, b)
}

func Benchmark_ParallelStabilize_withPreInitialize_16384(b *testing.B) {
	benchmarkParallelSize(16384, b)
}

func Benchmark_Stabilize_deep_2_32(b *testing.B) {
	benchmarkDepth(2, 32, b)
}

func Benchmark_Stabilize_deep_4_64(b *testing.B) {
	benchmarkDepth(4, 64, b)
}

func Benchmark_Stabilize_deep_4_96(b *testing.B) {
	benchmarkDepth(4, 96, b)
}

func Benchmark_Stabilize_deep_4_128(b *testing.B) {
	benchmarkDepth(4, 128, b)
}

func Benchmark_Stabilize_deep_8_256(b *testing.B) {
	benchmarkDepth(8, 256, b)
}

func Benchmark_Stabilize_deep_8_512(b *testing.B) {
	benchmarkDepth(8, 512, b)
}

func benchmarkSize(size int, b *testing.B) {
	ctx := testContext()
	nodes := make([]Incr[string], size)
	for x := 0; x < size; x++ {
		nodes[x] = Var(ctx, fmt.Sprintf("var_%d", x))
	}

	var cursor int
	for x := size; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Map2(ctx, nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	graph := New()
	_ = Observe(ctx, graph, nodes[len(nodes)-1])

	// this is what we care about
	b.ResetTimer()
	var err error
	for n := 0; n < b.N; n++ {
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		for x := 0; x < size>>1; x++ {
			graph.SetStale(nodes[rand.Intn(size)])
		}
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		for x := 0; x < size>>2; x++ {
			graph.SetStale(nodes[rand.Intn(size)])
		}
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkParallelSize(size int, b *testing.B) {
	ctx := testContext()
	nodes := make([]Incr[string], size)
	for x := 0; x < size; x++ {
		nodes[x] = Var(ctx, fmt.Sprintf("var_%d", x))
	}

	var cursor int
	for x := size; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Map2(ctx, nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	graph := New()
	_ = Observe(ctx, graph, nodes[0])

	// this is what we care about
	b.ResetTimer()
	var err error
	for n := 0; n < b.N; n++ {
		err = graph.ParallelStabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		for x := 0; x < size>>1; x++ {
			graph.SetStale(nodes[rand.Intn(size)])
		}
		err = graph.ParallelStabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		for x := 0; x < size>>2; x++ {
			graph.SetStale(nodes[rand.Intn(size)])
		}
		err = graph.ParallelStabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkDepth(width, depth int, b *testing.B) {
	ctx := testContext()

	vars := make([]VarIncr[string], width)
	for x := 0; x < width; x++ {
		vars[x] = Var(ctx, fmt.Sprintf("var_%d", x))
	}

	nodes := make([]Incr[string], width*depth)
	var nodeIndex int
	for y := 0; y < depth; y++ {
		for x := 0; x < width; x++ {
			if y == 0 {
				nodes[nodeIndex] = Map(ctx, vars[x], mapAppend(fmt.Sprintf("->%d", nodeIndex)))
			} else {
				previousIndex := ((y - 1) * width) + x
				nodes[nodeIndex] = Map(ctx, nodes[previousIndex], mapAppend(fmt.Sprintf("->%d", nodeIndex)))
			}
			nodeIndex++
		}
	}

	graph := New(
		GraphMaxRecomputeHeapHeight(512),
	)

	observers := make([]ObserveIncr[string], width)
	for x := 0; x < width; x++ {
		observers[x] = Observe(ctx, graph, nodes[(width*(depth-1))+x])
	}

	// this is what we care about
	b.ResetTimer()
	var err error
	for n := 0; n < b.N; n++ {
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}

		graph.SetStale(vars[rand.Intn(width)])

		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}
