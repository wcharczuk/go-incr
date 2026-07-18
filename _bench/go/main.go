// Command go-incr-bench runs the cross-library comparison benchmark suite for
// go-incr, emitting one JSON object per benchmark case to stdout.
//
// The OCaml counterpart in ../ocaml/bench.ml implements the identical set of
// cases against Jane Street's incremental, so results can be compared directly.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime/pprof"
	"time"

	incr "github.com/wcharczuk/go-incr"
)

var (
	flagFilter     = flag.String("filter", "", "only run cases whose name contains this substring")
	flagVerify     = flag.Bool("verify", false, "print observed values for each graph shape instead of benchmarking")
	flagCPUProfile = flag.String("cpuprofile", "", "write a CPU profile of the selected cases to this path")
	flagStats      = flag.Bool("stats", false, "report how many nodes each shape recomputes, instead of benchmarking")
	flagMemProfile = flag.String("memprofile", "", "write an allocation profile of the selected cases to this path")
)

// Timing parameters, kept identical to the OCaml harness in ../ocaml/bench.ml
// so that both libraries are measured the same way: grow a batch until it is
// long enough to dwarf clock resolution, then run batches until the total time
// and round count minimums are met.
const (
	minBatchSeconds = 0.01
	minTotalSeconds = 0.5
	minRounds       = 5
)

// result is the record emitted per benchmark case; the OCaml harness emits the
// same field names so the two outputs can be joined on (library, name).
type result struct {
	Library string  `json:"library"`
	Name    string  `json:"name"`
	Group   string  `json:"group"`
	Size    int     `json:"size"`
	Iters   int     `json:"iters"`
	NsPerOp float64 `json:"ns_per_op"`
	MinNs   float64 `json:"min_ns"`
}

func main() {
	flag.Parse()
	if *flagVerify {
		verify()
		return
	}
	if *flagStats {
		stats()
		return
	}
	if *flagCPUProfile != "" {
		f, err := os.Create(*flagCPUProfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "cpuprofile:", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintln(os.Stderr, "cpuprofile:", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}
	for _, c := range cases() {
		if *flagFilter != "" && !contains(c.name, *flagFilter) {
			continue
		}
		emit(run(c))
	}
	if *flagMemProfile != "" {
		f, err := os.Create(*flagMemProfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "memprofile:", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
			fmt.Fprintln(os.Stderr, "memprofile:", err)
			os.Exit(1)
		}
	}
}

// verify prints the observed value of each graph shape after a known mutation.
// The OCaml harness prints the same lines; if the two agree we know both
// libraries are building equivalent graphs and actually propagating, rather
// than one of them quietly stabilizing a graph with no necessary nodes.
func verify() {
	w := buildWide(1024, incr.NewSequentialIdentifierProvier(1))
	mustStabilize(w.g)
	fmt.Printf("wide/1024 initial=%d\n", w.obs.Value())
	w.vars[5].Set(999)
	mustStabilize(w.g)
	fmt.Printf("wide/1024 after_set=%d\n", w.obs.Value())

	d := buildDeep(128, incr.NewSequentialIdentifierProvier(1))
	mustStabilize(d.g)
	fmt.Printf("deep/128 initial=%d\n", d.obs.Value())
	d.v.Set(7)
	mustStabilize(d.g)
	fmt.Printf("deep/128 after_set=%d\n", d.obs.Value())

	g := newGraph(256)
	sel := incr.Var(g, 0)
	b := incr.Bind(g, sel, func(bs incr.Scope, which int) incr.Incr[int] {
		return buildChain(bs, which, 64)
	})
	bo := incr.MustObserve(g, b)
	mustStabilize(g)
	fmt.Printf("bind/64 initial=%d\n", bo.Value())
	sel.Set(3)
	mustStabilize(g)
	fmt.Printf("bind/64 after_set=%d\n", bo.Value())

	g2 := newGraph(1024)
	sel2 := incr.Var(g2, 0)
	binds := make([]incr.Incr[int], 256)
	for j := 0; j < 256; j++ {
		binds[j] = incr.Bind(g2, sel2, func(bs incr.Scope, which int) incr.Incr[int] {
			return incr.Map(bs, incr.Return(bs, which+j), func(x int) int { return x + 1 })
		})
	}
	wo := incr.MustObserve(g2, buildTreeReduce(g2, binds))
	mustStabilize(g2)
	fmt.Printf("bind_wide/256 initial=%d\n", wo.Value())
	sel2.Set(3)
	mustStabilize(g2)
	fmt.Printf("bind_wide/256 after_set=%d\n", wo.Value())
}

// stats reports, for each graph shape, how many nodes the initial stabilization
// recomputes and how many a single subsequent mutation recomputes.
//
// This is the algorithmic comparison rather than the performance one: the OCaml
// harness prints the same lines, and if both libraries recompute the same node
// counts for the same graph and the same mutation then they are performing
// equivalent work per stabilization. A higher count on either side means that
// side is missing a propagation optimization, independent of how fast its nodes
// happen to run.
func stats() {
	report := func(name string, g *incr.Graph, mutate func()) {
		e := incr.ExpertGraph(g)
		mustStabilize(g)
		initial := e.NumNodesRecomputed()
		initialDirect := e.NumNodesRecomputedDirectly()
		mutate()
		mustStabilize(g)
		fmt.Printf("%s initial_recomputed=%d update_recomputed=%d update_direct=%d total_nodes=%d\n",
			name,
			initial,
			e.NumNodesRecomputed()-initial,
			e.NumNodesRecomputedDirectly()-initialDirect,
			e.NumNodes(),
		)
	}

	w := buildWide(1024, incr.NewSequentialIdentifierProvier(1))
	report("wide/1024", w.g, func() { w.vars[5].Set(999) })

	for _, d := range []int{128, 2048} {
		dg := buildDeep(d, incr.NewSequentialIdentifierProvier(1))
		report(fmt.Sprintf("deep/%d", d), dg.g, func() { dg.v.Set(7) })
	}

	g := newGraph(256)
	sel := incr.Var(g, 0)
	incr.MustObserve(g, incr.Bind(g, sel, func(bs incr.Scope, which int) incr.Incr[int] {
		return buildChain(bs, which, 64)
	}))
	report("bind/swap_chain/64", g, func() { sel.Set(3) })

	g2 := newGraph(1024)
	sel2 := incr.Var(g2, 0)
	binds := make([]incr.Incr[int], 256)
	for j := 0; j < 256; j++ {
		binds[j] = incr.Bind(g2, sel2, func(bs incr.Scope, which int) incr.Incr[int] {
			return incr.Map(bs, incr.Return(bs, which+j), func(x int) int { return x + 1 })
		})
	}
	incr.MustObserve(g2, buildTreeReduce(g2, binds))
	report("bind/wide_swap/256", g2, func() { sel2.Set(3) })

	// setting a var to the value it already holds; a library with a default
	// cutoff recomputes nothing here.
	w2 := buildWide(1024, incr.NewSequentialIdentifierProvier(1))
	report("wide/1024/update_same", w2.g, func() { w2.vars[0].Set(0) })
}

// benchCase is a single measurable case. setup runs once outside the timed
// region and returns the operation to measure; for construction benchmarks the
// returned op does all the work itself and setup is a no-op.
type benchCase struct {
	name  string
	group string
	size  int
	setup func() func()
}

// timeBatch runs op batch times and returns the elapsed seconds.
func timeBatch(op func(), batch int) float64 {
	start := time.Now()
	for i := 0; i < batch; i++ {
		op()
	}
	return time.Since(start).Seconds()
}

// calibrate grows the batch size until a single timed batch is long enough that
// clock resolution is not a meaningful error term.
func calibrate(op func()) int {
	batch := 1
	for timeBatch(op, batch) < minBatchSeconds && batch < 1_000_000_000 {
		batch *= 2
	}
	return batch
}

func run(c benchCase) result {
	op := c.setup()

	// Warm up before calibrating so the batch size reflects steady state
	// rather than first-touch page faults and GC heap growth.
	for i := 0; i < 3; i++ {
		op()
	}

	batch := calibrate(op)
	rounds := 0
	total := 0.0
	best := math.Inf(1)
	for {
		elapsed := timeBatch(op, batch)
		total += elapsed
		if perOp := elapsed / float64(batch); perOp < best {
			best = perOp
		}
		rounds++
		if rounds >= minRounds && total >= minTotalSeconds {
			break
		}
	}
	iters := rounds * batch

	return result{
		Library: "go-incr",
		Name:    c.name,
		Group:   c.group,
		Size:    c.size,
		Iters:   iters,
		NsPerOp: total / float64(iters) * 1e9,
		MinNs:   best * 1e9,
	}
}

func emit(r result) {
	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(r); err != nil {
		fmt.Fprintln(os.Stderr, "encode:", err)
		os.Exit(1)
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

// newGraph builds a graph configured for benchmarking: a sequential identifier
// provider (the crypto/rand default dominates construction cost and is measured
// separately by the "ident_default" cases) and a height ceiling large enough for
// the deep-chain cases.
func newGraph(maxHeight int) *incr.Graph {
	return incr.New(
		incr.OptGraphMaxHeight(maxHeight),
		incr.OptGraphIdentifierProvider(incr.NewSequentialIdentifierProvier(1)),
	)
}

var background = context.Background()

func mustStabilize(g *incr.Graph) {
	if err := g.Stabilize(background); err != nil {
		panic(err)
	}
}

// buildTreeReduce reduces nodes pairwise with Map2 until a single root remains,
// giving a balanced tree of height log2(len(nodes)). A single leaf change
// therefore only dirties a logarithmic number of nodes, which is what makes
// this the interesting "wide graph" shape.
func buildTreeReduce(scope incr.Scope, nodes []incr.Incr[int]) incr.Incr[int] {
	add := func(a, b int) int { return a + b }
	for len(nodes) > 1 {
		next := make([]incr.Incr[int], 0, (len(nodes)+1)/2)
		i := 0
		for ; i+1 < len(nodes); i += 2 {
			next = append(next, incr.Map2(scope, nodes[i], nodes[i+1], add))
		}
		if i < len(nodes) {
			next = append(next, nodes[i])
		}
		nodes = next
	}
	return nodes[0]
}

// wideGraph is N leaf vars, each mapped, reduced by a balanced Map2 tree.
type wideGraph struct {
	g    *incr.Graph
	vars []incr.VarIncr[int]
	obs  incr.ObserveIncr[int]
}

// buildWideEqual is buildWide with inputs that ignore being set to the value they already
// hold, which is what OCaml incremental does by default. Without it the update_same case
// compares a library that cuts off against one that does not, which measures the semantic
// difference rather than the cost of the machinery.
func buildWideEqual(n int, ident incr.IdentifierProvider) *wideGraph {
	opts := []incr.GraphOption{incr.OptGraphMaxHeight(1024)}
	if ident != nil {
		opts = append(opts, incr.OptGraphIdentifierProvider(ident))
	}
	g := incr.New(opts...)

	vars := make([]incr.VarIncr[int], n)
	leaves := make([]incr.Incr[int], n)
	for i := 0; i < n; i++ {
		vars[i] = incr.VarEqual(g, i)
		leaves[i] = incr.Map(g, vars[i], func(v int) int { return v + 1 })
	}
	root := buildTreeReduce(g, leaves)
	obs := incr.MustObserve(g, root)
	return &wideGraph{g: g, vars: vars, obs: obs}
}

func buildWide(n int, ident incr.IdentifierProvider) *wideGraph {
	opts := []incr.GraphOption{incr.OptGraphMaxHeight(1024)}
	if ident != nil {
		opts = append(opts, incr.OptGraphIdentifierProvider(ident))
	}
	g := incr.New(opts...)

	vars := make([]incr.VarIncr[int], n)
	leaves := make([]incr.Incr[int], n)
	for i := 0; i < n; i++ {
		vars[i] = incr.Var(g, i)
		leaves[i] = incr.Map(g, vars[i], func(v int) int { return v + 1 })
	}
	root := buildTreeReduce(g, leaves)
	obs := incr.MustObserve(g, root)
	return &wideGraph{g: g, vars: vars, obs: obs}
}

// deepGraph is a single var feeding a chain of `depth` map nodes.
type deepGraph struct {
	g   *incr.Graph
	v   incr.VarIncr[int]
	obs incr.ObserveIncr[int]
}

func buildDeep(depth int, ident incr.IdentifierProvider) *deepGraph {
	opts := []incr.GraphOption{incr.OptGraphMaxHeight(depth + 64)}
	if ident != nil {
		opts = append(opts, incr.OptGraphIdentifierProvider(ident))
	}
	g := incr.New(opts...)

	v := incr.Var(g, 0)
	var cur incr.Incr[int] = v
	for i := 0; i < depth; i++ {
		cur = incr.Map(g, cur, func(x int) int { return x + 1 })
	}
	obs := incr.MustObserve(g, cur)
	return &deepGraph{g: g, v: v, obs: obs}
}

// buildChain builds a chain of `depth` maps rooted at a Return of `seed`,
// used as the body of bind subgraphs.
func buildChain(scope incr.Scope, seed, depth int) incr.Incr[int] {
	var cur incr.Incr[int] = incr.Return(scope, seed)
	for i := 0; i < depth; i++ {
		cur = incr.Map(scope, cur, func(x int) int { return x + 1 })
	}
	return cur
}

func cases() []benchCase {
	var out []benchCase

	// --- Wide graphs -------------------------------------------------------
	for _, n := range []int{1024, 16384} {

		// Full construction plus the initial stabilization that computes
		// every node for the first time.
		out = append(out, benchCase{
			name:  fmt.Sprintf("wide/construct/%d", n),
			group: "wide", size: n,
			setup: func() func() {
				return func() {
					w := buildWide(n, incr.NewSequentialIdentifierProvier(1))
					mustStabilize(w.g)
				}
			},
		})

		// Same as above but on the library's default crypto/rand identifier
		// provider, to quantify what that default costs.
		out = append(out, benchCase{
			name:  fmt.Sprintf("wide/construct_ident_default/%d", n),
			group: "wide_ident", size: n,
			setup: func() func() {
				return func() {
					w := buildWide(n, nil)
					mustStabilize(w.g)
				}
			},
		})

		// The core incremental metric: change one leaf, restabilize. Only
		// log2(n) nodes should recompute.
		out = append(out, benchCase{
			name:  fmt.Sprintf("wide/update_one/%d", n),
			group: "wide", size: n,
			setup: func() func() {
				w := buildWide(n, incr.NewSequentialIdentifierProvier(1))
				mustStabilize(w.g)
				i := 0
				return func() {
					i++
					w.vars[i%n].Set(i)
					mustStabilize(w.g)
				}
			},
		})

		// Change every leaf, restabilize: the worst case, every node dirty.
		out = append(out, benchCase{
			name:  fmt.Sprintf("wide/update_all/%d", n),
			group: "wide", size: n,
			setup: func() func() {
				w := buildWide(n, incr.NewSequentialIdentifierProvier(1))
				mustStabilize(w.g)
				i := 0
				return func() {
					i++
					for j := 0; j < n; j++ {
						w.vars[j].Set(i + j)
					}
					mustStabilize(w.g)
				}
			},
		})

		// Set a leaf to the value it already holds. Jane Street's incremental
		// cuts off propagation on an unchanged value by default; go-incr does
		// not, so this case measures that semantic difference.
		out = append(out, benchCase{
			name:  fmt.Sprintf("wide/update_same/%d", n),
			group: "wide_cutoff", size: n,
			setup: func() func() {
				w := buildWide(n, incr.NewSequentialIdentifierProvier(1))
				mustStabilize(w.g)
				return func() {
					w.vars[0].Set(0)
					mustStabilize(w.g)
				}
			},
		})

		// The same case with an input that cuts off, which is the comparable
		// configuration: what update_same above measures against OCaml is mostly the
		// default, not the machinery.
		out = append(out, benchCase{
			name:  fmt.Sprintf("wide/update_same_equal/%d", n),
			group: "wide_cutoff", size: n,
			setup: func() func() {
				w := buildWideEqual(n, incr.NewSequentialIdentifierProvier(1))
				mustStabilize(w.g)
				return func() {
					w.vars[0].Set(0)
					mustStabilize(w.g)
				}
			},
		})
	}

	// --- Deep graphs ------------------------------------------------------
	// Depth 1 is included to isolate the fixed per-stabilization overhead from
	// the marginal per-node recompute cost that the deeper sizes measure.
	for _, d := range []int{1, 128, 2048} {

		out = append(out, benchCase{
			name:  fmt.Sprintf("deep/construct/%d", d),
			group: "deep", size: d,
			setup: func() func() {
				return func() {
					dg := buildDeep(d, incr.NewSequentialIdentifierProvier(1))
					mustStabilize(dg.g)
				}
			},
		})

		// Change the root: the whole chain must recompute in height order,
		// which is the pure recompute-heap throughput measurement.
		out = append(out, benchCase{
			name:  fmt.Sprintf("deep/update_one/%d", d),
			group: "deep", size: d,
			setup: func() func() {
				dg := buildDeep(d, incr.NewSequentialIdentifierProvier(1))
				mustStabilize(dg.g)
				i := 0
				return func() {
					i++
					dg.v.Set(i)
					mustStabilize(dg.g)
				}
			},
		})
	}

	// --- Bind graphs ------------------------------------------------------
	// Toggling the bind's left-hand side tears down the previous subgraph and
	// builds a fresh one of `depth` nodes, exercising scope tracking, height
	// adjustment and invalidation.
	for _, d := range []int{64, 512} {
		out = append(out, benchCase{
			name:  fmt.Sprintf("bind/swap_chain/%d", d),
			group: "bind", size: d,
			setup: func() func() {
				g := newGraph(d + 128)
				sel := incr.Var(g, 0)
				b := incr.Bind(g, sel, func(bs incr.Scope, which int) incr.Incr[int] {
					return buildChain(bs, which, d)
				})
				incr.MustObserve(g, b)
				mustStabilize(g)
				i := 0
				return func() {
					i++
					sel.Set(i)
					mustStabilize(g)
				}
			},
		})
	}

	// Many independent binds sharing one left-hand side, summed by a tree.
	// A single set therefore rebuilds n subgraphs at once, which is the
	// "large bind graph" stress case.
	for _, n := range []int{256, 4096} {
		out = append(out, benchCase{
			name:  fmt.Sprintf("bind/wide_swap/%d", n),
			group: "bind_wide", size: n,
			setup: func() func() {
				g := newGraph(1024)
				sel := incr.Var(g, 0)
				binds := make([]incr.Incr[int], n)
				for j := 0; j < n; j++ {
					binds[j] = incr.Bind(g, sel, func(bs incr.Scope, which int) incr.Incr[int] {
						return incr.Map(bs, incr.Return(bs, which+j), func(x int) int { return x + 1 })
					})
				}
				incr.MustObserve(g, buildTreeReduce(g, binds))
				mustStabilize(g)
				i := 0
				return func() {
					i++
					sel.Set(i)
					mustStabilize(g)
				}
			},
		})
	}

	// A bind whose subgraph is built from the node kinds beyond map/map2, so that
	// their input-list handling is actually exercised. There is no OCaml
	// counterpart for this case; it exists to A/B go-incr against itself.
	for _, d := range []int{128} {
		out = append(out, benchCase{
			name:  fmt.Sprintf("bind/swap_mixed/%d", d),
			group: "bind_mixed", size: d,
			setup: func() func() {
				g := newGraph(8 * d)
				sel := incr.Var(g, 0)
				b := incr.Bind(g, sel, func(bs incr.Scope, which int) incr.Incr[int] {
					// d independent map3 groups, each over three distinct inputs
					// and wrapped in a cutoff, reduced by a map2 tree. map3 fills
					// its input list at construction and cutoff fills its on
					// demand, so this covers both of the shapes that changed.
					//
					// The inputs are deliberately distinct: feeding one node into
					// several slots of the same multi-input node makes bind
					// rebuilds blow up superlinearly, which is a pre-existing
					// issue unrelated to what this case is meant to measure.
					groups := make([]incr.Incr[int], d)
					for j := 0; j < d; j++ {
						a := incr.Return(bs, which+j)
						b2 := incr.Map(bs, a, func(x int) int { return x + 1 })
						c := incr.Map(bs, b2, func(x int) int { return x * 2 })
						m := incr.Map3(bs, a, b2, c, func(x, y, z int) int { return x + y - z })
						groups[j] = incr.Cutoff(bs, m, func(prev, next int) bool { return false })
					}
					return buildTreeReduce(bs, groups)
				})
				incr.MustObserve(g, b)
				mustStabilize(g)
				i := 0
				return func() {
					i++
					sel.Set(i)
					mustStabilize(g)
				}
			},
		})
	}

	// Construction cost of a bind-heavy graph, including the initial
	// stabilization that first materializes every subgraph.
	for _, n := range []int{4096} {
		out = append(out, benchCase{
			name:  fmt.Sprintf("bind/wide_construct/%d", n),
			group: "bind_wide", size: n,
			setup: func() func() {
				return func() {
					g := newGraph(1024)
					sel := incr.Var(g, 0)
					binds := make([]incr.Incr[int], n)
					for j := 0; j < n; j++ {
						binds[j] = incr.Bind(g, sel, func(bs incr.Scope, which int) incr.Incr[int] {
							return incr.Map(bs, incr.Return(bs, which+j), func(x int) int { return x + 1 })
						})
					}
					incr.MustObserve(g, buildTreeReduce(g, binds))
					mustStabilize(g)
				}
			},
		})
	}

	return out
}
