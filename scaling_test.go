package incr

import (
	"context"
	"fmt"
	"math"
	"os"
	"testing"
	"time"
)

// This file guards against pathological scaling rather than against slowness.
//
// A constant factor costs every user a little; a superlinear cost makes the
// library unusable past some graph size, and which size that is depends on the
// shape rather than on anything a user can see. Several such cliffs have already
// been found here -- an O(block) search in the recompute heap, a min-height scan
// that restarted from zero, a teardown that re-walked its subgraph once per
// duplicate edge, and a bind relink that leaked one edge per rebuild -- so each
// shape below is measured at several sizes and its growth exponent asserted.
//
// The exponent is log(t2/t1) / log(n2/n1): 1.0 is linear, ~1.1 is n log n, 2.0 is
// quadratic. Thresholds are deliberately loose, because these are wall-clock
// measurements in a test and the point is to catch a cliff, not to police a few
// percent. A shape that has gone quadratic will read close to 2 and clear any
// threshold set for linear-ish behavior.

// scalingCase measures one graph shape at a size and returns the per-operation
// cost. Setup is excluded; only the operation under test is timed.
type scalingCase struct {
	name string
	// sizes must be increasing, and are usually powers of two so the exponent is
	// computed over clean ratios.
	sizes []int
	// maxExponent is the growth exponent above which the shape is considered
	// pathological.
	maxExponent float64
	measure     func(t *testing.T, size int) time.Duration
}

// exponent fits the growth of cost against size across all adjacent size pairs
// and returns the largest, so a cliff that only appears at the top end is not
// averaged away.
func exponent(sizes []int, costs []time.Duration) (worst float64, detail string) {
	worst = math.Inf(-1)
	for i := 1; i < len(sizes); i++ {
		sizeRatio := float64(sizes[i]) / float64(sizes[i-1])
		costRatio := float64(costs[i]) / float64(costs[i-1])
		if costRatio <= 0 {
			continue
		}
		e := math.Log(costRatio) / math.Log(sizeRatio)
		detail += fmt.Sprintf("  %6d -> %6d: %12v -> %12v  exponent %.2f\n",
			sizes[i-1], sizes[i], costs[i-1], costs[i], e)
		if e > worst {
			worst = e
		}
	}
	return
}

// timeOp returns a per-operation cost, taking the best of several rounds.
//
// The minimum rather than the mean, and several rounds rather than one, because
// these run on shared machines where an unlucky round can be several times slow. A
// spike inflates a mean but cannot lower a minimum, so best-of-N rejects exactly the
// noise that made an earlier version of these tests flaky in CI.
func timeOp(op func()) time.Duration {
	op()
	iters := 1
	for {
		start := time.Now()
		for i := 0; i < iters; i++ {
			op()
		}
		if time.Since(start) > 5*time.Millisecond || iters >= 1<<20 {
			break
		}
		iters *= 4
	}
	best := time.Duration(math.MaxInt64)
	for range 5 {
		start := time.Now()
		for i := 0; i < iters; i++ {
			op()
		}
		if per := time.Since(start) / time.Duration(iters); per < best {
			best = per
		}
	}
	return best
}

// skipTimingUnlessRequested skips a wall-clock measurement unless it was asked for.
//
// These are opt-in rather than opt-out, which is a deliberate inversion of the usual
// -short convention. They measure elapsed time, so they can fail for reasons that have
// nothing to do with the code: run inside the full suite they compete with a large heap
// and with whatever GC the preceding two hundred tests left behind, and a threshold loose
// enough to survive that is too loose to catch anything. Widening them until they stop
// flaking would leave a test that passes regardless.
//
// What the invariants actually rest on is the deterministic work-counting tests, which
// run always: Test_operators_work, Test_Reduce_work, Test_Join_work,
// Test_MapValues_onlyChangedKeys and the node counts in duplicate_edge_test.go. Those
// assert the same asymptotic claims by counting callbacks, so they cannot flake.
//
// Set INCR_SCALING_TESTS=1 to run these, which CI does in a dedicated step where nothing
// else is competing; `make scaling` does the same.
func skipTimingUnlessRequested(t *testing.T) {
	t.Helper()
	if os.Getenv("INCR_SCALING_TESTS") == "" {
		t.Skip("wall-clock scaling measurement; set INCR_SCALING_TESTS=1 to run")
	}
}

func runScalingCase(t *testing.T, c scalingCase) {
	t.Helper()
	skipTimingUnlessRequested(t)
	costs := make([]time.Duration, len(c.sizes))
	for i, size := range c.sizes {
		costs[i] = c.measure(t, size)
	}
	worst, detail := exponent(c.sizes, costs)
	t.Logf("%s: worst exponent %.2f (limit %.2f)\n%s", c.name, worst, c.maxExponent, detail)
	if worst > c.maxExponent {
		t.Errorf("%s scales superlinearly: exponent %.2f exceeds %.2f\n%s",
			c.name, worst, c.maxExponent, detail)
	}
}

// buildFanIn returns n vars feeding a balanced map2 reduction tree.
func buildFanIn(g *Graph, n int) []VarIncr[int] {
	vars := make([]VarIncr[int], n)
	nodes := make([]Incr[int], n)
	for i := 0; i < n; i++ {
		vars[i] = Var(g, i)
		nodes[i] = vars[i]
	}
	for len(nodes) > 1 {
		next := make([]Incr[int], 0, (len(nodes)+1)/2)
		var i int
		for ; i+1 < len(nodes); i += 2 {
			next = append(next, Map2(g, nodes[i], nodes[i+1], func(a, b int) int { return a + b }))
		}
		if i < len(nodes) {
			next = append(next, nodes[i])
		}
		nodes = next
	}
	MustObserve(g, nodes[0])
	return vars
}

// buildFanOut returns one var with n dependents, reduced so all are necessary.
//
// This is the shape where a single node accumulates a large dependent list, which
// is what makes an O(list) edge removal pathological.
func buildFanOut(g *Graph, n int) (VarIncr[int], ObserveIncr[int]) {
	v := Var(g, 1)
	leaves := make([]Incr[int], n)
	for i := 0; i < n; i++ {
		leaves[i] = Map(g, v, func(x int) int { return x + 1 })
	}
	nodes := leaves
	for len(nodes) > 1 {
		next := make([]Incr[int], 0, (len(nodes)+1)/2)
		var i int
		for ; i+1 < len(nodes); i += 2 {
			next = append(next, Map2(g, nodes[i], nodes[i+1], func(a, b int) int { return a + b }))
		}
		if i < len(nodes) {
			next = append(next, nodes[i])
		}
		nodes = next
	}
	return v, MustObserve(g, nodes[0])
}

func Test_scaling_construct(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:  "construct fan-in + first stabilize",
		sizes: []int{1024, 4096, 16384},
		// construction is linear in nodes; allow headroom for allocator and GC
		// effects that grow mildly with heap size.
		maxExponent: 1.45,
		measure: func(t *testing.T, size int) time.Duration {
			return timeOp(func() {
				g := New(OptGraphMaxHeight(1024), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
				_ = buildFanIn(g, size)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		},
	})
}

func Test_scaling_updateOne(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:  "single leaf update in fan-in tree",
		sizes: []int{1024, 4096, 16384},
		// only log2(n) nodes recompute, so cost should barely move with n; this is
		// mostly a check that it does not become linear or worse.
		maxExponent: 0.6,
		measure: func(t *testing.T, size int) time.Duration {
			g := New(OptGraphMaxHeight(1024), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
			vars := buildFanIn(g, size)
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			var i int
			return timeOp(func() {
				i++
				vars[i%size].Set(i)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		},
	})
}

func Test_scaling_updateAll(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:  "every leaf updated in fan-in tree",
		sizes: []int{1024, 4096, 16384},
		// This case is linear by nature -- it changes every leaf -- and the lower size
		// step measures 0.95 to 1.02 accordingly. The top step reads higher and varies
		// between runs, because at 16384 leaves the graph no longer fits in cache and
		// the per-node cost rises for reasons that are not algorithmic. The limit has to
		// clear that while staying below quadratic, which is what a real regression
		// here would look like.
		maxExponent: 1.7,
		measure: func(t *testing.T, size int) time.Duration {
			g := New(OptGraphMaxHeight(1024), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
			vars := buildFanIn(g, size)
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			var i int
			return timeOp(func() {
				i++
				for j := 0; j < size; j++ {
					vars[j].Set(i + j)
				}
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		},
	})
}

func Test_scaling_deepChain(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:        "root update through a deep chain",
		sizes:       []int{256, 1024, 4096},
		maxExponent: 1.35,
		measure: func(t *testing.T, size int) time.Duration {
			g := New(OptGraphMaxHeight(size+64), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
			v := Var(g, 0)
			var cur Incr[int] = v
			for i := 0; i < size; i++ {
				cur = Map(g, cur, func(x int) int { return x + 1 })
			}
			MustObserve(g, cur)
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			var i int
			return timeOp(func() {
				i++
				v.Set(i)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		},
	})
}

// Test_scaling_fanOutTeardown is the shape that exposes an O(dependents) edge
// removal: one node with many dependents, torn down by unobserving.
func Test_scaling_fanOutTeardown(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:        "teardown of a wide fan-out",
		sizes:       []int{512, 2048, 8192},
		maxExponent: 1.45,
		measure: func(t *testing.T, size int) time.Duration {
			return timeOp(func() {
				g := New(OptGraphMaxHeight(1024), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
				_, o := buildFanOut(g, size)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
				o.Unobserve(ctx)
				if n := ExpertGraph(g).NumNodes(); n != 0 {
					t.Fatalf("expected a fully torn down graph, have %d nodes", n)
				}
			})
		},
	})
}

// Test_scaling_bindSharedControl is the nested-bind shape reduced to its
// essential feature: many binds sharing one input, so that one node accumulates a
// large dependent list which every rebuild has to edit.
func Test_scaling_bindSharedControl(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:        "rebuild of many binds sharing one control var",
		sizes:       []int{256, 1024, 4096},
		maxExponent: 1.45,
		measure: func(t *testing.T, size int) time.Duration {
			g := New(OptGraphMaxHeight(1024), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
			sel := Var(g, 0)
			binds := make([]Incr[int], size)
			for j := 0; j < size; j++ {
				binds[j] = Bind(g, sel, func(bs Scope, which int) Incr[int] {
					return Map(bs, Return(bs, which+j), func(x int) int { return x + 1 })
				})
			}
			nodes := binds
			for len(nodes) > 1 {
				next := make([]Incr[int], 0, (len(nodes)+1)/2)
				var i int
				for ; i+1 < len(nodes); i += 2 {
					next = append(next, Map2(g, nodes[i], nodes[i+1], func(a, b int) int { return a + b }))
				}
				if i < len(nodes) {
					next = append(next, nodes[i])
				}
				nodes = next
			}
			MustObserve(g, nodes[0])
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			var i int
			return timeOp(func() {
				i++
				sel.Set(i)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		},
	})
}

// Test_scaling_bindDeepSubgraph rebuilds a bind whose right-hand side is a deep
// chain, so the whole subgraph is discarded and rebuilt each pass.
func Test_scaling_bindDeepSubgraph(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:        "rebuild of a deep bind subgraph",
		sizes:       []int{128, 512, 2048},
		maxExponent: 1.45,
		measure: func(t *testing.T, size int) time.Duration {
			g := New(OptGraphMaxHeight(4*size+128), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
			sel := Var(g, 0)
			b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
				var cur Incr[int] = Return(bs, which)
				for i := 0; i < size; i++ {
					cur = Map(bs, cur, func(x int) int { return x + 1 })
				}
				return cur
			})
			MustObserve(g, b)
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			var i int
			return timeOp(func() {
				i++
				sel.Set(i)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		},
	})
}

// Test_scaling_duplicateInputTeardown covers the arity^depth teardown that
// duplicate edges used to cause.
func Test_scaling_duplicateInputTeardown(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:        "rebuild of a chain of duplicate-input nodes",
		sizes:       []int{64, 256, 1024},
		maxExponent: 1.45,
		measure: func(t *testing.T, size int) time.Duration {
			g := New(OptGraphMaxHeight(4*size+128), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
			sel := Var(g, 0)
			b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
				var cur Incr[int] = Return(bs, which)
				for i := 0; i < size; i++ {
					cur = Map3(bs, cur, cur, cur, func(a, b, c int) int { return a + b - c })
				}
				return cur
			})
			MustObserve(g, b)
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			var i int
			return timeOp(func() {
				i++
				sel.Set(i)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		},
	})
}

// Test_scaling_mapNTeardown is the mirror of the fan-out case on the input side:
// one node with a large input list, where removing each edge in turn would be
// quadratic in the number of inputs.
func Test_scaling_mapNTeardown(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:        "teardown of a wide mapN",
		sizes:       []int{512, 2048, 8192},
		maxExponent: 1.45,
		measure: func(t *testing.T, size int) time.Duration {
			return timeOp(func() {
				g := New(OptGraphMaxHeight(256), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
				inputs := make([]Incr[int], size)
				for i := 0; i < size; i++ {
					inputs[i] = Var(g, i)
				}
				m := MapN(g, func(vs ...int) int {
					var out int
					for _, v := range vs {
						out += v
					}
					return out
				}, inputs...)
				o := MustObserve(g, m)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
				o.Unobserve(ctx)
			})
		},
	})
}

// Test_scaling_mapNRemoveInputs covers removing a wide mapN's inputs one at a
// time.
//
// This case is quadratic and is expected to be, which is why its limit is set
// above 2 rather than near 1. A mapN's inputs are ordered -- the values reach the
// fold function in input order -- so removing one shifts the rest, and n removals
// move O(n^2) elements however the entry is found. Unlike every other case here
// this is not automatic work a user runs into: it takes an explicit loop over
// [MapNIncr.RemoveInput].
//
// The threshold still guards the shape: it catches this going cubic, or the
// ordered removal being joined by a second linear factor. Making it linear needs
// either permission to reorder inputs, or a tombstoned representation that keeps
// order without shifting; both are design decisions rather than fixes. See
// _bench/ALGORITHMS.md.
func Test_scaling_mapNRemoveInputs(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:        "removing every input of a wide mapN (known quadratic)",
		sizes:       []int{256, 1024, 4096},
		maxExponent: 2.35,
		measure: func(t *testing.T, size int) time.Duration {
			return timeOp(func() {
				g := New(OptGraphMaxHeight(256), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
				vars := make([]VarIncr[int], size)
				inputs := make([]Incr[int], size)
				for i := 0; i < size; i++ {
					vars[i] = Var(g, i)
					inputs[i] = vars[i]
				}
				m := MapN(g, func(vs ...int) int { return len(vs) }, inputs...)
				MustObserve(g, m)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
				for i := 0; i < size; i++ {
					if err := m.RemoveInput(vars[i].Node().ID()); err != nil {
						t.Fatal(err)
					}
				}
			})
		},
	})
}

// Test_scaling_manyObservers covers a node watched by many observers, and
// unobserving them one at a time.
func Test_scaling_manyObservers(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:        "unobserving many observers of one node",
		sizes:       []int{256, 1024, 4096},
		maxExponent: 1.55,
		measure: func(t *testing.T, size int) time.Duration {
			return timeOp(func() {
				g := New(OptGraphMaxHeight(256), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
				v := Var(g, 1)
				m := Map(g, v, func(x int) int { return x + 1 })
				obs := make([]ObserveIncr[int], size)
				for i := 0; i < size; i++ {
					obs[i] = MustObserve(g, m)
				}
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
				for _, o := range obs {
					o.Unobserve(ctx)
				}
			})
		},
	})
}

// Test_scaling_sparseHeights covers a graph whose nodes sit at widely separated
// heights under a large height ceiling, which is what the recompute heap's scan
// for its next non-empty height block has to cope with.
func Test_scaling_sparseHeights(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:        "updates under a large sparse height ceiling",
		sizes:       []int{1024, 4096, 16384},
		maxExponent: 0.6,
		measure: func(t *testing.T, size int) time.Duration {
			// a short chain, but a ceiling far above it, so any scan proportional
			// to the ceiling rather than to occupied heights shows up here.
			g := New(OptGraphMaxHeight(size), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
			v := Var(g, 0)
			var cur Incr[int] = v
			for i := 0; i < 8; i++ {
				cur = Map(g, cur, func(x int) int { return x + 1 })
			}
			MustObserve(g, cur)
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			var i int
			return timeOp(func() {
				i++
				v.Set(i)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		},
	})
}

// Test_scaling_aggregation compares the three ways of aggregating many inputs.
//
// This is the case the combinators exist for. [MapN] hands every value to its
// function on every pass, so a single input changing is O(inputs); a user who
// reaches for it to sum a few thousand values gets a graph that degrades with
// size for no visible reason. [ReduceBalanced] costs O(log n) and
// [UnorderedArrayFold] O(1), and these assertions are what hold them to that.
func Test_scaling_aggregation(t *testing.T) {
	ctx := context.Background()
	sizes := []int{256, 1024, 4096}

	// build returns a graph whose root aggregates size vars, along with the vars.
	type builder func(g *Graph, inputs []Incr[int]) Incr[int]
	setup := func(build builder) func(t *testing.T, size int) time.Duration {
		return func(t *testing.T, size int) time.Duration {
			g := New(OptGraphMaxHeight(1024), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
			vars := make([]VarIncr[int], size)
			inputs := make([]Incr[int], size)
			for i := 0; i < size; i++ {
				vars[i] = Var(g, i)
				inputs[i] = vars[i]
			}
			MustObserve(g, build(g, inputs))
			if err := g.Stabilize(ctx); err != nil {
				t.Fatal(err)
			}
			var i int
			return timeOp(func() {
				i++
				vars[i%size].Set(i)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		}
	}

	// MapN is linear by construction; asserted so that the comparison below is
	// anchored rather than assumed.
	runScalingCase(t, scalingCase{
		name:        "one input changes, aggregated with MapN (linear by design)",
		sizes:       sizes,
		maxExponent: 1.35,
		measure: setup(func(g *Graph, inputs []Incr[int]) Incr[int] {
			return MapN(g, func(vs ...int) int {
				var out int
				for _, v := range vs {
					out += v
				}
				return out
			}, inputs...)
		}),
	})

	// a balanced tree recomputes only the path from the changed leaf to the root
	runScalingCase(t, scalingCase{
		name:        "one input changes, aggregated with ReduceBalanced",
		sizes:       sizes,
		maxExponent: 0.5,
		measure: setup(func(g *Graph, inputs []Incr[int]) Incr[int] {
			return ReduceBalanced(g, func(a, b int) int { return a + b }, inputs...)
		}),
	})

	// an unordered fold adjusts its accumulator and touches nothing else
	runScalingCase(t, scalingCase{
		name:        "one input changes, aggregated with UnorderedArrayFold",
		sizes:       sizes,
		maxExponent: 0.35,
		measure: setup(func(g *Graph, inputs []Incr[int]) Incr[int] {
			return UnorderedArrayFold(g, 0,
				func(acc, v int) int { return acc + v },
				func(acc, old, new int) int { return acc - old + new },
				inputs...)
		}),
	})
}

// Test_scaling_repeatedRebuild covers cost per rebuild staying flat as a
// long-running process rebuilds the same bind over and over, which is where a
// per-rebuild leak shows up.
func Test_scaling_repeatedRebuild(t *testing.T) {
	ctx := context.Background()
	runScalingCase(t, scalingCase{
		name:  "cost per rebuild after N previous rebuilds",
		sizes: []int{500, 2000, 8000},
		// the "size" here is elapsed history, not graph size: cost per rebuild
		// should not grow with it at all.
		maxExponent: 0.35,
		measure: func(t *testing.T, size int) time.Duration {
			g := New(OptGraphMaxHeight(256), OptGraphIdentifierProvider(NewSequentialIdentifierProvier(1)))
			sel := Var(g, 0)
			b := Bind(g, sel, func(bs Scope, which int) Incr[int] {
				return Map(bs, Return(bs, which), func(x int) int { return x + 1 })
			})
			MustObserve(g, b)
			for i := 0; i < size; i++ {
				sel.Set(i)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			}
			i := size
			return timeOp(func() {
				i++
				sel.Set(i)
				if err := g.Stabilize(ctx); err != nil {
					t.Fatal(err)
				}
			})
		},
	})
}
