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
	ctx := context.Background()
	g := incr.New()
	nodes := make([]incr.Incr[string], SIZE)
	vars := make([]incr.VarIncr[string], 0, SIZE)
	for x := 0; x < SIZE; x++ {
		v := incr.Var(g, fmt.Sprintf("var_%d", x))
		vars = append(vars, v)
		nodes[x] = v
	}

	var cursor int
	for x := SIZE; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := incr.Map2(g, nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}

	if os.Getenv("DEBUG") != "" {
		ctx = incr.WithTracing(ctx)
	}
	_ = incr.Observe(g, nodes[0])

	var err error
	for n := 0; n < ROUNDS; n++ {
		err = g.Stabilize(ctx)
		if err != nil {
			fatal(err)
		}
		vars[rand.Intn(len(vars))].Set(fmt.Sprintf("set_%d", n))
		err = g.Stabilize(ctx)
		if err != nil {
			fatal(err)
		}
	}

	buf := new(bytes.Buffer)
	_ = incr.Dot(buf, g)
	fmt.Print(buf.String())
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%+v\n", err)
	os.Exit(1)
}
