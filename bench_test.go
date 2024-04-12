package incr

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
)

func Benchmark_createGraph_512(b *testing.B) {
	benchmarkCreateGraph(512, false, b)
}

func Benchmark_createGraph_preallocateNodes_512(b *testing.B) {
	benchmarkCreateGraph(512, true, b)
}

func Benchmark_createGraph_customIdentifierProvider_512(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(512, false, b)
}

func Benchmark_createGraph_preallocateNodes_customIdentifierProvider_512(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(512, true, b)
}

func Benchmark_createGraph_1024(b *testing.B) {
	benchmarkCreateGraph(1024, false, b)
}

func Benchmark_createGraph_preallocateNodes_1024(b *testing.B) {
	benchmarkCreateGraph(1024, true, b)
}

func Benchmark_createGraph_customIdentifierProvider_1024(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(1024, false, b)
}

func Benchmark_createGraph_preallocateNodes_customIdentifierProvider_1024(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(1024, true, b)
}

func Benchmark_createGraph_2048(b *testing.B) {
	benchmarkCreateGraph(2048, false, b)
}

func Benchmark_createGraph_preallocateNodes_2048(b *testing.B) {
	benchmarkCreateGraph(2048, true, b)
}

func Benchmark_createGraph_customIdentifierProvider_2048(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(2048, false, b)
}

func Benchmark_createGraph_preallocateNodes_customIdentifierProvider_2048(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(2048, true, b)
}

func Benchmark_createGraph_4096(b *testing.B) {
	benchmarkCreateGraph(4096, false, b)
}

func Benchmark_createGraph_preallocateNodes_4096(b *testing.B) {
	benchmarkCreateGraph(4096, true, b)
}

func Benchmark_createGraph_customIdentifierProvider_4096(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(4096, false, b)
}

func Benchmark_createGraphpreallocateNodes__customIdentifierProvider_4096(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(4096, true, b)
}

func Benchmark_createGraph_8192(b *testing.B) {
	benchmarkCreateGraph(8192, false, b)
}

func Benchmark_createGraph_preallocateNodes_8192(b *testing.B) {
	benchmarkCreateGraph(8192, true, b)
}

func Benchmark_createGraph_customIdentifierProvider_8192(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(8192, false, b)
}

func Benchmark_createGraphpreallocateNodes__customIdentifierProvider_8192(b *testing.B) {
	b.Cleanup(func() {
		SetIdentifierProvider(cryptoRandIdentifierProvider)
	})
	SetIdentifierProvider(counterIdentifierProvider)
	benchmarkCreateGraph(8192, true, b)
}

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

func Benchmark_Stabilize_recombinant_64(b *testing.B) {
	benchmarkRecombinantSize(64, b)
}

func Benchmark_Stabilize_recombinant_128(b *testing.B) {
	benchmarkRecombinantSize(128, b)
}

func Benchmark_Stabilize_recombinant_256(b *testing.B) {
	benchmarkRecombinantSize(256, b)
}

func Benchmark_Stabilize_recombinant_512(b *testing.B) {
	benchmarkRecombinantSize(512, b)
}

func Benchmark_ParallelStabilize_recombinant_64(b *testing.B) {
	benchmarkParallelRecombinantSize(64, b)
}

func Benchmark_ParallelStabilize_recombinant_128(b *testing.B) {
	benchmarkParallelRecombinantSize(128, b)
}

func Benchmark_ParallelStabilize_recombinant_256(b *testing.B) {
	benchmarkParallelRecombinantSize(256, b)
}

func Benchmark_ParallelStabilize_recombinant_512(b *testing.B) {
	benchmarkParallelRecombinantSize(512, b)
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

func Benchmark_Stabilize_nestedBinds_16(b *testing.B) {
	benchmarkNestedBinds(16, b)
}

func Benchmark_Stabilize_nestedBinds_32(b *testing.B) {
	benchmarkNestedBinds(32, b)
}

func Benchmark_Stabilize_nestedBinds_64(b *testing.B) {
	benchmarkNestedBinds(64, b)
}

func Benchmark_Stabilize_nestedBinds_128(b *testing.B) {
	benchmarkNestedBinds(128, b)
}

func longer(a, b *string) *string {
	if a == nil && b == nil {
		return nil
	}
	if a != nil && b == nil {
		return a
	}
	if a == nil && b != nil {
		return b
	}
	if len(*a) > len(*b) {
		return a
	}
	return b
}

func ref[A any](v A) *A { return &v }

func makeBenchmarkGraph(size int, preallocate bool) (*Graph, []Incr[*string]) {
	var options []GraphOption
	if preallocate {
		options = append(options, OptGraphPreallocateNodesSize(size<<1))
	}
	graph := New(options...)
	nodes := make([]Incr[*string], size)
	for x := 0; x < size; x++ {
		nodes[x] = Var(graph, ref(fmt.Sprintf("var_%d", x)))
	}

	var cursor int
	for x := size; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Map2(graph, nodes[cursor+y], nodes[cursor+y+1], longer)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	_ = MustObserve(graph, nodes[len(nodes)-1])
	return graph, nodes
}

func makeBenchmarkRecombinantGraph(size int) (*Graph, VarIncr[*string], ObserveIncr[*string]) {
	g := New()

	input := Var(g, ref("input"))
	input.Node().SetLabel("input")
	nodes := []Incr[*string]{input}
	m00 := Map(g, input, ident)
	m00.Node().SetLabel("exp-m0-0")
	m01 := Map(g, input, ident)
	m01.Node().SetLabel("exp-m0-1")
	nodes = append(nodes, m01)

	cursor := 1
	level := 1
	for x := 2; x <= size; x <<= 1 {
		var levelNode int
		for y := 0; y < x; y++ {
			m00 := Map(g, nodes[cursor+y], ident)
			m00.Node().SetLabel(fmt.Sprintf("exp-m%d-%d", level, levelNode))
			nodes = append(nodes, m00)
			levelNode++
			m01 := Map(g, nodes[cursor+y], ident)
			m01.Node().SetLabel(fmt.Sprintf("exp-m%d-%d", level, levelNode))
			nodes = append(nodes, m01)
			levelNode++
		}
		cursor += x
		level++
	}
	cursor = len(nodes) - 1
	for x := size << 1; x > 0; x >>= 1 {
		var levelNode int
		for y := 0; y < x-1; y += 2 {
			m := Map2(g, nodes[cursor-y], nodes[cursor-(y+1)], longer)
			m.Node().SetLabel(fmt.Sprintf("cont-m%d-%d", level, levelNode))
			nodes = append(nodes, m)
			levelNode++
		}
		cursor += (x >> 1)
		level--
	}
	observer := MustObserve(g, nodes[len(nodes)-1])
	return g, input, observer
}

func benchmarkCreateGraph(size int, preallocate bool, b *testing.B) {
	for x := 0; x < b.N; x++ {
		_, _ = makeBenchmarkGraph(size, preallocate)
	}
}

func benchmarkSize(size int, b *testing.B) {
	graph, nodes := makeBenchmarkGraph(size, false /*preallocate*/)
	ctx := context.Background()
	b.ResetTimer()
	var err error
	for n := 0; n < b.N; n++ {
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		graph.SetStale(nodes[rand.Intn(size)])
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		graph.SetStale(nodes[rand.Intn(size)])
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkParallelSize(size int, b *testing.B) {
	graph, nodes := makeBenchmarkGraph(size, false /*preallocate*/)
	ctx := testContext()
	b.ResetTimer()
	var err error
	for n := 0; n < b.N; n++ {
		err = graph.ParallelStabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		graph.SetStale(nodes[rand.Intn(size)])
		err = graph.ParallelStabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		graph.SetStale(nodes[rand.Intn(size)])
		err = graph.ParallelStabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkRecombinantSize(size int, b *testing.B) {
	graph, input, observer := makeBenchmarkRecombinantGraph(size)
	ctx := testContext()
	b.ResetTimer()
	var err error
	for n := 0; n < b.N; n++ {
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if observer.Value() == nil {
			b.Fail()
		}
		graph.SetStale(input)
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if observer.Value() == nil {
			b.Fail()
		}
		graph.SetStale(input)
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if observer.Value() == nil {
			b.Fail()
		}
	}
}

func benchmarkParallelRecombinantSize(size int, b *testing.B) {
	graph, input, observer := makeBenchmarkRecombinantGraph(size)
	ctx := testContext()
	b.ResetTimer()
	var err error
	for n := 0; n < b.N; n++ {
		err = graph.ParallelStabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if observer.Value() == nil {
			b.Fail()
		}
		graph.SetStale(input)
		err = graph.ParallelStabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if observer.Value() == nil {
			b.Fail()
		}
		graph.SetStale(input)
		err = graph.ParallelStabilize(ctx)
		if err != nil {
			b.Fatal(err)
		}
		if observer.Value() == nil {
			b.Fail()
		}
	}
}

func benchmarkDepth(width, depth int, b *testing.B) {
	graph := New(
		OptGraphMaxHeight(1024),
	)
	vars := make([]VarIncr[string], width)
	for x := 0; x < width; x++ {
		vars[x] = Var(graph, fmt.Sprintf("var_%d", x))
	}
	nodes := make([]Incr[string], width*depth)
	var nodeIndex int
	for y := 0; y < depth; y++ {
		for x := 0; x < width; x++ {
			if y == 0 {
				nodes[nodeIndex] = Map(graph, vars[x], mapAppend(fmt.Sprintf("->%d", nodeIndex)))
			} else {
				previousIndex := ((y - 1) * width) + x
				nodes[nodeIndex] = Map(graph, nodes[previousIndex], mapAppend(fmt.Sprintf("->%d", nodeIndex)))
			}
			nodeIndex++
		}
	}
	observers := make([]ObserveIncr[string], width)
	for x := 0; x < width; x++ {
		observers[x] = MustObserve(graph, nodes[(width*(depth-1))+x])
	}
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
	ctx := testContext()
	graph := New(
		OptGraphMaxHeight(2048),
	)
	bindControl := Var(graph, 1)
	o := makeNestedBindGraph(graph, depth, bindControl)
	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		err := graph.Stabilize(ctx)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if o.Value() == 0 {
			b.Errorf("value is unset")
			b.FailNow()
		}
		bindControl.Set(2)
		err = graph.Stabilize(ctx)
		if err != nil {
			b.Error(err)
			b.FailNow()
		}
		if o.Value() == 0 {
			b.Errorf("value is unset")
			b.FailNow()
		}
	}
}

func makeNestedBindGraph(g *Graph, depth int, bindControl VarIncr[int]) ObserveIncr[int] {
	vars := make([]VarIncr[int], 0, depth)
	for x := 0; x < depth; x++ {
		vars = append(vars, Var(g, x))
	}
	binds := make([]BindIncr[int], 0, depth*depth)
	final := make([]Incr[int], 0, depth)
	for y := 0; y < depth; y++ {
		for x := 0; x < depth; x++ {
			switch y {
			case 0:
				b := Bind(g, bindControl, func(x, _ int) BindFunc[int, int] {
					return func(_ Scope, which int) Incr[int] {
						return vars[(x+which)%depth]
					}
				}(x, y))
				binds = append(binds, b)
			case depth - 1:
				b := Bind(g, bindControl, func(x, y int) BindFunc[int, int] {
					return func(_ Scope, which int) Incr[int] {
						bindIndex := ((y - 1) * depth) + (x+which)%depth
						return binds[bindIndex]
					}
				}(x, y))
				final = append(final, b)
			default:
				b := Bind(g, bindControl, func(x, y int) BindFunc[int, int] {
					return func(_ Scope, which int) Incr[int] {
						bindIndex := ((y - 1) * depth) + (x+which)%depth
						return binds[bindIndex]
					}
				}(x, y))
				binds = append(binds, b)
			}
		}
	}
	m := MapN(g, func(values ...int) (out int) {
		for _, v := range values {
			out += v
		}
		return
	}, final...)
	om := MustObserve(g, m)
	return om
}
