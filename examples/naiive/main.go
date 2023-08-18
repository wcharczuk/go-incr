package main

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"github.com/wcharczuk/go-incr"
)

const (
	SIZE   = 8192
	ROUNDS = 100
)

type INode[A any] interface {
	Value(context.Context) A
}

type Node[A, B any] struct {
	Children []INode[A]
	Action   func(context.Context, ...A) B
}

func (n Node[A, B]) Value(ctx context.Context) B {
	inputs := make([]A, len(n.Children))
	for x := 0; x < len(n.Children); x++ {
		inputs[x] = n.Children[x].Value(ctx)
	}
	return n.Action(ctx, inputs...)
}

func Var[A any](v A) INode[A] {
	return &Node[any, A]{
		Action: func(ctx context.Context, _ ...any) A {
			return v
		},
	}
}

func Map[A, B any](child0, child1 INode[A], fn func(context.Context, ...A) B) INode[B] {
	return &Node[A, B]{
		Children: []INode[A]{child0, child1},
		Action:   fn,
	}
}

func concat(_ context.Context, values ...string) string {
	return strings.Join(values, "")
}

func main() {
	naiiveVars, naiiveNodes := makeNodes()
	var naiiveResults []time.Duration
	ctx := context.TODO()
	for n := 0; n < ROUNDS; n++ {
		start := time.Now()
		randomNode := naiiveVars[rand.Intn(len(naiiveVars))].(*Node[any, string])
		randomNode.Action = func(ctx context.Context, _ ...any) string {
			return fmt.Sprintf("set_%d", n)
		}
		_ = naiiveNodes[len(naiiveNodes)-1].Value(ctx)
		naiiveResults = append(naiiveResults, time.Since(start))
	}

	incrVars, incrNodes := makeIncrNodes()
	graph := incr.New(incrNodes[0])
	var incrResults []time.Duration
	for n := 0; n < ROUNDS; n++ {
		start := time.Now()
		incrVars[rand.Intn(len(incrVars))].Set(fmt.Sprintf("set_%d", n))
		_ = graph.ParallelStabilize(ctx)
		incrResults = append(incrResults, time.Since(start))
	}

	fmt.Println("results!")
	fmt.Printf("naiive: %v\n", avgDurations(naiiveResults).Round(time.Microsecond))
	fmt.Printf("incr: %v\n", avgDurations(incrResults).Round(time.Microsecond))
}

func avgDurations(values []time.Duration) time.Duration {
	accum := new(big.Int)
	for _, v := range values {
		accum.Add(accum, big.NewInt(int64(v)))
	}
	return time.Duration(accum.Div(accum, big.NewInt(int64(len(values)))).Int64())
}

func makeNodes() (vars []INode[string], nodes []INode[string]) {
	nodes = make([]INode[string], SIZE)
	vars = make([]INode[string], SIZE)
	for x := 0; x < SIZE; x++ {
		v := Var(fmt.Sprintf("var_%d", x))
		nodes[x] = v
		vars[x] = v
	}

	var cursor int
	for x := SIZE; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := Map[string, string](nodes[cursor+y], nodes[cursor+y+1], concat)
			nodes = append(nodes, n)
		}
		cursor += x
	}
	return
}

func concat2(a, b string) string {
	return a + b
}

func makeIncrNodes() (vars []incr.VarIncr[string], nodes []incr.Incr[string]) {
	nodes = make([]incr.Incr[string], SIZE)
	vars = make([]incr.VarIncr[string], SIZE)
	for x := 0; x < SIZE; x++ {
		v := incr.Var(fmt.Sprintf("var_%d", x))
		vars[x] = v
		nodes[x] = v
	}

	var cursor int
	for x := SIZE; x > 0; x >>= 1 {
		for y := 0; y < x-1; y += 2 {
			n := incr.Map2(nodes[cursor+y], nodes[cursor+y+1], concat2)
			nodes = append(nodes, n)
		}
		cursor += x
	}
	return
}
