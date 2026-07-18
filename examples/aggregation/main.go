// Command aggregation compares the ways of aggregating many inputs, on the same workload,
// and reports what each one actually costs.
//
// The scenario is a portfolio: several thousand positions, and a few numbers derived from
// all of them that have to stay current as individual positions are marked. Every
// approach below produces the same answers. They differ only in what a single position
// changing costs, and that difference is the difference between a graph that scales and
// one that does not.
//
// The numbers this prints are the reason the package documentation leads with choosing a
// combinator rather than with the list of them.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wcharczuk/go-incr"
)

const positions = 4096

func main() {
	fmt.Printf("a portfolio of %d positions; one position is marked, then the derived\n", positions)
	fmt.Printf("values are brought back up to date. cost is per single mark.\n\n")

	fmt.Printf("%-24s %-46s %12s %10s\n", "approach", "what it does on one change", "per mark", "fn calls")
	fmt.Println(dashes(96))

	// MapN hands every value to its function on every pass, so marking one position
	// re-reads all of them. This is the one people reach for first, because it is the
	// obvious shape: "a function of all my inputs".
	measure("MapN", "reads all 4096 values", func(g *incr.Graph, vars []incr.VarIncr[int], calls *int) incr.INode {
		inputs := asIncrs(vars)
		return incr.MustObserve(g, incr.MapN(g, func(values ...int) int {
			total := 0
			for _, v := range values {
				*calls++
				total += v
			}
			return total
		}, inputs...))
	})

	// ArrayFold is the same cost with a more natural shape; it is here to make the point
	// that the ergonomics and the cost are separate choices.
	measure("ArrayFold", "reads all 4096 values", func(g *incr.Graph, vars []incr.VarIncr[int], calls *int) incr.INode {
		inputs := asIncrs(vars)
		return incr.MustObserve(g, incr.ArrayFold(g, 0, func(acc, v int) int {
			*calls++
			return acc + v
		}, inputs...))
	})

	// ReduceBalanced combines pairwise through a balanced tree, so only the path from the
	// marked position up to the root recomputes: twelve combines rather than 4096 reads.
	// It needs associativity and nothing else, which is what makes it the right answer
	// for a maximum or a concatenation.
	measure("ReduceBalanced", "recombines log2(4096) = 12 nodes", func(g *incr.Graph, vars []incr.VarIncr[int], calls *int) incr.INode {
		inputs := asIncrs(vars)
		return incr.MustObserve(g, incr.ReduceBalanced(g, func(a, b int) int {
			*calls++
			return a + b
		}, inputs...))
	})

	// UnorderedArrayFold keeps a running total and adjusts it from the marked position's
	// old and new values. That is one subtraction and one addition however large the
	// portfolio, but it only works because addition has an inverse.
	measure("UnorderedArrayFold", "adjusts the total: -old +new", func(g *incr.Graph, vars []incr.VarIncr[int], calls *int) incr.INode {
		inputs := asIncrs(vars)
		return incr.MustObserve(g, incr.UnorderedArrayFold(g, 0,
			func(acc, v int) int { return acc + v },
			func(acc, old, next int) int {
				*calls++
				return acc - old + next
			}, inputs...))
	})

	fmt.Println()
	fmt.Println("the same choice for an aggregate with no inverse")
	fmt.Println(dashes(96))
	fmt.Println("A maximum cannot be maintained by adjustment: withdrawing the current largest")
	fmt.Println("position tells you nothing about the next largest, so there is no update rule to")
	fmt.Println("write. ReduceBalanced still applies, and is the reason it exists alongside the fold.")
	fmt.Println()

	measure("ReduceBalanced (max)", "recombines 12 nodes", func(g *incr.Graph, vars []incr.VarIncr[int], calls *int) incr.INode {
		inputs := asIncrs(vars)
		return incr.MustObserve(g, incr.ReduceBalanced(g, func(a, b int) int {
			*calls++
			return max(a, b)
		}, inputs...))
	})

	fmt.Println()
	fmt.Println("every approach gives the same total; only the cost of staying current differs.")
}

// measure builds a graph with the given aggregate, marks one position repeatedly, and
// reports the per-mark cost and how many times the aggregate's function was called.
func measure(
	name, describe string,
	build func(g *incr.Graph, vars []incr.VarIncr[int], calls *int) incr.INode,
) {
	g := incr.New(incr.OptGraphMaxHeight(1024),
		incr.OptGraphIdentifierProvider(incr.NewSequentialIdentifierProvider(1)))

	vars := make([]incr.VarIncr[int], positions)
	for i := range vars {
		vars[i] = incr.Var(g, 100+i)
	}
	var calls int
	build(g, vars, &calls)

	if err := g.Stabilize(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}

	// the first pass necessarily touches everything, so measure the steady state
	calls = 0
	const marks = 200
	start := time.Now()
	for i := range marks {
		vars[i%positions].Set(500 + i)
		if err := g.Stabilize(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "%+v\n", err)
			os.Exit(1)
		}
	}
	elapsed := time.Since(start) / marks

	fmt.Printf("%-24s %-46s %12s %10d\n", name, describe, elapsed.String(), calls/marks)
}

func asIncrs(vars []incr.VarIncr[int]) []incr.Incr[int] {
	out := make([]incr.Incr[int], len(vars))
	for i, v := range vars {
		out[i] = v
	}
	return out
}

func dashes(n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = '-'
	}
	return string(out)
}
