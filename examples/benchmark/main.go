package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/wcharczuk/go-incr"
)

const (
	SIZE   = 2048
	ROUNDS = 1000
)

func concat(a, b string) string {
	return a + b
}

func main() {
	rand.Seed(time.Now().Unix())

	nodes := make([]incr.Incr[string], SIZE)
	for x := 0; x < SIZE; x++ {
		nodes[x] = incr.Var(fmt.Sprintf("var_%d", x))
	}

	var cursor int
	for x := SIZE; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := incr.Apply2(nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	ctx := context.Background()
	if os.Getenv("DEBUG") != "" {
		ctx = incr.WithTracing(ctx)
	}

	gs := nodes[len(nodes)-1]
	incr.Initialize(ctx, gs)

	var err error
	for n := 0; n < ROUNDS; n++ {
		err = incr.Stabilize(ctx, gs)
		if err != nil {
			fatal(err)
		}
		for x := 0; x < SIZE>>1; x++ {
			incr.SetStale(nodes[rand.Intn(SIZE)])
		}
		err = incr.Stabilize(ctx, gs)
		if err != nil {
			fatal(err)
		}
		for x := 0; x < SIZE>>2; x++ {
			incr.SetStale(nodes[rand.Intn(SIZE)])
		}
		err = incr.Stabilize(ctx, gs)
		if err != nil {
			fatal(err)
		}
	}

	stats := incr.GraphStats(nodes[0])
	fmt.Println("nodes            :", stats.Nodes())
	fmt.Println("nodes recomputed :", stats.NodesRecomputed())
	fmt.Println("nodes changed    :", stats.NodesChanged())
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%+v\n", err)
	os.Exit(1)
}
