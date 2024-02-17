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

func Benchmark_closureCall(b *testing.B) {
	benchmarkClosure(b)
}

func Benchmark_directCall(b *testing.B) {
	benchmarkDirectCall(b)
}

func makeBenchmarkGraph(size int) (*Graph, []Incr[string]) {
	graph := New()
	nodes := make([]Incr[string], size)
	for x := 0; x < size; x++ {
		nodes[x] = Var(graph, fmt.Sprintf("var_%d", x))
	}

	var cursor int
	for x := size; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Map2(graph, nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	_ = MustObserve(graph, nodes[len(nodes)-1])
	return graph, nodes
}

func benchmarkSize(size int, b *testing.B) {
	graph, nodes := makeBenchmarkGraph(size)
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
	graph, nodes := makeBenchmarkGraph(size)
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
				b := Bind(g, bindControl, func(x, y int) BindFunc[int, int] {
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

func benchmarkClosure(b *testing.B) {
	ctx := testContext()
	g := New()
	r0 := Return(g, 1)
	r1 := Return(g, 2)
	m := &map2Incr[int, int, int]{
		n:       NewNode("map2"),
		a:       r0,
		b:       r1,
		fn:      func(_ context.Context, a, b int) (int, error) { return a + b, nil },
		parents: []INode{r0, r1},
	}
	m.n.createdIn = g
	_ = MustObserve(g, m)

	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		_ = m.Node().stabilizeFn(ctx)
	}
}

func benchmarkDirectCall(b *testing.B) {
	ctx := testContext()
	g := New()
	r0 := Return(g, 1)
	r1 := Return(g, 2)
	m := &map2Incr[int, int, int]{
		n:       NewNode("map2"),
		a:       r0,
		b:       r1,
		fn:      func(_ context.Context, a, b int) (int, error) { return a + b, nil },
		parents: []INode{r0, r1},
	}
	m.n.createdIn = g
	_ = MustObserve(g, m)

	b.ResetTimer()
	for x := 0; x < b.N; x++ {
		_ = m.Stabilize(ctx)
	}
}
