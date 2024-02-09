package incr

import (
	"context"
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

func Benchmark_Stabilize_nestedBinds_64(b *testing.B) {
	benchmarkNestedBinds(64, b)
}

func Benchmark_Stabilize_nestedBinds_128(b *testing.B) {
	benchmarkNestedBinds(128, b)
}

func Benchmark_Stabilize_nestedBinds_256(b *testing.B) {
	benchmarkNestedBinds(256, b)
}

func Benchmark_Stabilize_nestedBinds_512(b *testing.B) {
	benchmarkNestedBinds(512, b)
}

func Benchmark_Stabilize_nestedBinds_1024(b *testing.B) {
	benchmarkNestedBinds(1024, b)
}

func Benchmark_Stabilize_connectedGraph_with_nestedBinds_64(b *testing.B) {
	benchmarkConnectedGraphWithNestedBinds(64, b)
}

func Benchmark_Stabilize_connectedGraph_with_nestedBinds_128(b *testing.B) {
	benchmarkConnectedGraphWithNestedBinds(128, b)
}

// func Benchmark_Stabilize_connectedGraph_with_nestedBinds_256(b *testing.B) {
// 	benchmarkConnectedGraphWithNestedBinds(256, b)
// }

// func Benchmark_Stabilize_connectedGraph_with_nestedBinds_512(b *testing.B) {
// 	benchmarkConnectedGraphWithNestedBinds(512, b)
// }

func benchmarkSize(size int, b *testing.B) {
	nodes := make([]Incr[string], size)
	for x := 0; x < size; x++ {
		nodes[x] = Var(Root(), fmt.Sprintf("var_%d", x))
	}

	var cursor int
	for x := size; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Map2(Root(), nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	graph := New()
	_ = Observe(Root(), graph, nodes[len(nodes)-1])

	// this is what we care about
	ctx := context.Background()
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
	nodes := make([]Incr[string], size)
	for x := 0; x < size; x++ {
		nodes[x] = Var(Root(), fmt.Sprintf("var_%d", x))
	}

	var cursor int
	for x := size; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Map2(Root(), nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	graph := New()
	_ = Observe(Root(), graph, nodes[0])

	// this is what we care about
	ctx := context.Background()
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
	vars := make([]VarIncr[string], width)
	for x := 0; x < width; x++ {
		vars[x] = Var(Root(), fmt.Sprintf("var_%d", x))
	}

	nodes := make([]Incr[string], width*depth)
	var nodeIndex int
	for y := 0; y < depth; y++ {
		for x := 0; x < width; x++ {
			if y == 0 {
				nodes[nodeIndex] = Map(Root(), vars[x], mapAppend(fmt.Sprintf("->%d", nodeIndex)))
			} else {
				previousIndex := ((y - 1) * width) + x
				nodes[nodeIndex] = Map(Root(), nodes[previousIndex], mapAppend(fmt.Sprintf("->%d", nodeIndex)))
			}
			nodeIndex++
		}
	}
	graph := New(
		GraphMaxRecomputeHeapHeight(1024),
	)
	observers := make([]ObserveIncr[string], width)
	for x := 0; x < width; x++ {
		observers[x] = Observe(Root(), graph, nodes[(width*(depth-1))+x])
	}

	// this is what we care about
	ctx := context.Background()
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

func benchmarkNestedBinds(depth int, b *testing.B) {
	ctx := context.Background()
	fakeFormula := Var(Root(), "fakeFormula")
	g, o := makeNestedBindGraph(depth, fakeFormula)
	for x := 0; x < b.N; x++ {
		err := g.Stabilize(ctx)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if o.Value() == nil {
			b.FailNow()
		}
		g.SetStale(fakeFormula)
		err = g.Stabilize(ctx)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
	}
}

func makeNestedBindGraph(depth int, fakeFormula VarIncr[string]) (*Graph, ObserveIncr[*int]) {
	graph := New(
		GraphMaxRecomputeHeapHeight(1024),
	)
	cache := make(map[string]Incr[*int])
	var m func(bs *BindScope, t int) Incr[*int]
	left_bound := 3
	right_bound := 9
	m = func(bs *BindScope, t int) Incr[*int] {
		key := fmt.Sprintf("m-%d", t)
		if _, ok := cache[key]; ok {
			return WithinBindScope(bs, cache[key])
		}
		r := Bind(bs, fakeFormula, func(bs *BindScope, formula string) Incr[*int] {
			if t == 0 {
				out := 0
				r := Return(bs, &out)
				r.Node().SetLabel("m-0")
				return r
			}
			var bindOutput Incr[*int]
			offset := 1
			if t >= left_bound && t < right_bound {
				li := m(bs, t-1)
				bindOutput = Map2(bs, li, Return(bs, &offset), func(l *int, r *int) *int {
					if l == nil || r == nil {
						return nil
					}
					out := *l + *r
					return &out
				})
			} else {
				bindOutput = m(bs, t-1)
			}
			bindOutput.Node().SetLabel(fmt.Sprintf("%s-output", key))
			return bindOutput
		})

		r.Node().SetLabel(fmt.Sprintf("m(%d)", t))
		cache[key] = r
		return r
	}
	o := m(Root(), depth)
	return graph, Observe(Root(), graph, o)
}

func benchmarkConnectedGraphWithNestedBinds(depth int, b *testing.B) {
	ctx := context.Background()
	fakeFormula := Var(Root(), "fakeFormula")
	observed := make([]ObserveIncr[*int], depth)
	g := New(
		GraphMaxRecomputeHeapHeight(1024),
	)
	for i := 0; i < depth; i++ {
		o := makeSimpleNestedBindGraph(g, depth, fakeFormula)
		observed[i] = o
	}
	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		err := g.Stabilize(ctx)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		for _, o := range observed {
			if o.Value() == nil {
				b.FailNow()
			}
		}
	}
}

func makeSimpleNestedBindGraph(graph *Graph, depth int, fakeFormula VarIncr[string]) ObserveIncr[*int] {
	cache := make(map[string]Incr[*int])
	var f func(bs *BindScope, t int) Incr[*int]
	f = func(bs *BindScope, t int) Incr[*int] {
		key := fmt.Sprintf("f-%d", t)
		if _, ok := cache[key]; ok {
			return WithinBindScope(bs, cache[key])
		}
		r := Bind(bs, fakeFormula, func(bs *BindScope, formula string) Incr[*int] {
			if t <= 0 {
				out := 0
				r := Return(bs, &out)
				r.Node().SetLabel("f-0")
				return r
			}
			return Map(bs, f(bs, t-1), func(r *int) *int {
				out := *r
				return &out
			})
		})
		r.Node().SetLabel(fmt.Sprintf("f(%d)", t))
		cache[key] = r
		return r
	}
	o := f(Root(), depth)
	return Observe(Root(), graph, o)
}
