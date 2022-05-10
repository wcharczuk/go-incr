package incr

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
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

func benchmarkSize(size int, b *testing.B) {
	rand.Seed(time.Now().Unix())

	nodes := make([]Incr[string], size)
	for x := 0; x < size; x++ {
		nodes[x] = Var(fmt.Sprintf("var_%d", x))
	}

	var cursor int
	for x := size; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Apply2(nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	gs := nodes[len(nodes)-1]
	Initialize(testContext(), gs)
	for n := 0; n < b.N; n++ {
		_ = Stabilize(testContext(), gs)
		for x := 0; x < size>>2; x++ {
			SetStale(nodes[rand.Intn(size)])
		}
		_ = Stabilize(testContext(), gs)
		for x := 0; x < size>>2; x++ {
			SetStale(nodes[rand.Intn(size)])
		}
		_ = Stabilize(testContext(), gs)
	}
}

func concat(_ context.Context, a, b string) (string, error) {
	return a + b, nil
}
