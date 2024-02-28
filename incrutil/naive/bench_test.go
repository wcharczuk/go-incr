package naive

import (
	"fmt"
	"testing"
)

func Benchmark_512(b *testing.B) {
	benchmarkSize(512, b)
}

func Benchmark_1024(b *testing.B) {
	benchmarkSize(1024, b)
}

func Benchmark_2048(b *testing.B) {
	benchmarkSize(2048, b)
}

func Benchmark_4096(b *testing.B) {
	benchmarkSize(4096, b)
}

func Benchmark_8192(b *testing.B) {
	benchmarkSize(8192, b)
}

func Benchmark_16384(b *testing.B) {
	benchmarkSize(16384, b)
}

func benchmarkSize(size int, b *testing.B) {
	root, _ := makeBenchmarkGraph(size)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = root.Value()
		_ = root.Value()
		_ = root.Value()
	}
}

func ref[A any](v A) *A { return &v }

func longer(values ...*string) (longest *string) {
	for _, v := range values {
		if v == nil {
			continue
		}
		if longest == nil {
			longest = v
		}
		if len(*v) > len(*longest) {
			longest = v
		}
	}
	return
}

func makeBenchmarkGraph(size int) (Node[*string], []Node[*string]) {
	nodes := make([]Node[*string], size)
	for x := 0; x < size; x++ {
		nodes[x] = Var(ref(fmt.Sprintf("var_%d", x)))
	}
	var cursor int
	for x := size; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Map(longer, nodes[cursor+y], nodes[cursor+y+1])
			nodes = append(nodes, n)
		}
		cursor += x
	}
	return nodes[len(nodes)-1], nodes
}
