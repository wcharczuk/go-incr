package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"

	"github.com/wcharczuk/go-incr"
)

const (
	SIZE   = 32
	ROUNDS = 3
)

func concat(a, b string) string {
	return a + b
}

func main() {
	nodes := make([]incr.Incr[string], SIZE)
	vars := make([]incr.VarIncr[string], 0, SIZE)
	for x := 0; x < SIZE; x++ {
		v := incr.Var(fmt.Sprintf("var_%d", x))
		vars = append(vars, v)
		nodes[x] = v
	}

	var cursor int
	for x := SIZE; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := incr.Map2(nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	graph := incr.New()

	ctx := context.Background()
	if os.Getenv("DEBUG") != "" {
		ctx = incr.WithTracing(ctx)
	}

	_ = incr.Observe(graph, nodes[0])

	var err error
	for n := 0; n < ROUNDS; n++ {
		err = graph.Stabilize(ctx)
		if err != nil {
			fatal(err)
		}
		vars[rand.Intn(len(vars))].Set(fmt.Sprintf("set_%d", n))
		err = graph.Stabilize(ctx)
		if err != nil {
			fatal(err)
		}
	}

	buf := new(bytes.Buffer)
	_ = incr.Dot(buf, nodes[0])
	fmt.Print(buf.String())
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%+v\n", err)
	os.Exit(1)
}
