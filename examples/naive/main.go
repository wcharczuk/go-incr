package main

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/naive"
)

const (
	SIZE   = 8192 << 3
	ROUNDS = 256
)

func concatN(values ...string) string {
	return strings.Join(values, "")
}

func main() {
	ctx := context.Background()
	naiveVars, naiveNodes := makeNaiveNodes()
	var naiveResults []time.Duration
	for n := 0; n < ROUNDS; n++ {
		start := time.Now()
		randomVar := naiveVars[rand.Intn(len(naiveVars))]
		randomVar.SetValue(fmt.Sprintf("set_%d", n))
		_ = naiveNodes[len(naiveNodes)-1].Value()
		naiveResults = append(naiveResults, time.Since(start))
	}

	graph := incr.New()
	incrVars, incrNodes := makeIncrNodes(ctx, graph)
	incr.MustObserve(graph, incrNodes[0])

	var incrResults []time.Duration
	for n := 0; n < ROUNDS; n++ {
		start := time.Now()
		incrVars[rand.Intn(len(incrVars))].Set(fmt.Sprintf("set_%d", n))
		_ = graph.Stabilize(ctx)
		incrResults = append(incrResults, time.Since(start))
	}

	fmt.Println("results!")
	fmt.Printf("naive: %v\n", avgDurations(naiveResults).Round(time.Microsecond))
	fmt.Printf("incr: %v\n", avgDurations(incrResults).Round(time.Microsecond))
}

func avgDurations(values []time.Duration) time.Duration {
	accum := new(big.Int)
	for _, v := range values {
		accum.Add(accum, big.NewInt(int64(v)))
	}
	return time.Duration(accum.Div(accum, big.NewInt(int64(len(values)))).Int64())
}

func makeNaiveNodes() (vars []naive.VarNode[string], nodes []naive.Node[string]) {
	nodes = make([]naive.Node[string], SIZE)
	vars = make([]naive.VarNode[string], SIZE)
	for x := 0; x < SIZE; x++ {
		v := naive.Var(fmt.Sprintf("var_%d", x))
		nodes[x] = v
		vars[x] = v
	}

	var cursor int
	for x := SIZE; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := naive.Map[string, string](concatN, nodes[cursor+y], nodes[cursor+y+1])
			nodes = append(nodes, n)
		}
		cursor += x
	}
	return
}

func makeIncrNodes(_ context.Context, graph *incr.Graph) (vars []incr.VarIncr[string], nodes []incr.Incr[string]) {
	nodes = make([]incr.Incr[string], SIZE)
	vars = make([]incr.VarIncr[string], SIZE)
	for x := 0; x < SIZE; x++ {
		v := incr.Var(graph, fmt.Sprintf("var_%d", x))
		vars[x] = v
		nodes[x] = v
	}

	var cursor int
	for x := SIZE; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := incr.Map2(graph, nodes[cursor+y], nodes[cursor+y+1], func(a, b string) string {
				return concatN(a, b)
			})
			nodes = append(nodes, n)
		}
		cursor += x
	}
	return
}
