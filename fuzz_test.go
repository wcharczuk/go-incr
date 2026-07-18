package incr

import (
	"context"
	"testing"
)

// FuzzGraph drives a graph through arbitrary sequences of construction, observation and
// stabilization, and checks two things after every step: that the graph's structure is
// internally consistent, and that unobserving everything drains it completely.
//
// Both properties are the ones that failed in practice. Every structural bug found in this
// library so far -- a bind that relinked without removing the other side of the edge, a
// duplicated input torn down twice, a node count that underflowed -- was invisible to a
// test that only checked values, because values stayed correct while the bookkeeping
// beneath them drifted. They showed up only under a shape nobody had thought to write
// down, which is what a fuzzer is for.
//
// Operations are allowed to fail. A random program can legitimately exceed the height
// limit or observe something already observed, and rejecting those is correct behavior;
// only an inconsistent graph or a leak is a failure.
func FuzzGraph(f *testing.F) {
	// a var, a map over it, an observe, a set, a stabilize
	f.Add([]byte{0, 0, 1, 0, 4, 0, 6, 0, 7, 0})
	// a bind, rebuilt several times
	f.Add([]byte{0, 0, 3, 0, 4, 0, 7, 0, 6, 1, 7, 0, 6, 2, 7, 0})
	// map2 over the same node twice, which is the duplicate-input shape
	f.Add([]byte{0, 0, 2, 0, 4, 0, 7, 0})
	// several observers over one node
	f.Add([]byte{0, 0, 1, 0, 4, 0, 4, 0, 4, 0, 7, 0, 5, 0, 7, 0})
	// a node that panics, observed and stabilized, then released
	f.Add([]byte{0, 1, 8, 0, 4, 0, 7, 0, 7, 0, 5, 0, 7, 0})

	f.Fuzz(func(t *testing.T, program []byte) {
		const (
			maxNodes = 24
			maxOps   = 96
		)
		ctx := context.Background()
		g := New(OptGraphMaxHeight(512))

		var nodes []Incr[int]
		var vars []VarIncr[int]
		var observers []ObserveIncr[int]

		// pick indexes into the pools deterministically from the program byte
		pickNode := func(b byte) Incr[int] { return nodes[int(b)%len(nodes)] }

		ops := 0
		for i := 0; i+1 < len(program) && ops < maxOps; i += 2 {
			op, arg := program[i]%9, program[i+1]
			ops++

			switch op {
			case 0: // a new var
				if len(nodes) >= maxNodes {
					continue
				}
				v := Var(g, int(arg))
				vars = append(vars, v)
				nodes = append(nodes, v)

			case 1: // map over an existing node
				if len(nodes) == 0 || len(nodes) >= maxNodes {
					continue
				}
				nodes = append(nodes, Map(g, pickNode(arg), func(x int) int { return x + 1 }))

			case 2: // map2 over two existing nodes, which may well be the same node twice
				if len(nodes) == 0 || len(nodes) >= maxNodes {
					continue
				}
				a, b := pickNode(arg), pickNode(arg/3)
				nodes = append(nodes, Map2(g, a, b, func(x, y int) int { return x + y }))

			case 3: // a bind, which builds nodes in its own scope and replaces them
				if len(nodes) == 0 || len(nodes) >= maxNodes {
					continue
				}
				nodes = append(nodes, Bind(g, pickNode(arg), func(bs Scope, which int) Incr[int] {
					inner := Return(bs, which)
					if which%2 == 0 {
						return Map(bs, inner, func(x int) int { return x * 2 })
					}
					return Map2(bs, inner, Return(bs, 1), func(x, y int) int { return x + y })
				}))

			case 4: // observe
				if len(nodes) == 0 {
					continue
				}
				o, err := Observe(g, pickNode(arg))
				if err != nil {
					continue
				}
				observers = append(observers, o)

			case 5: // unobserve
				if len(observers) == 0 {
					continue
				}
				index := int(arg) % len(observers)
				observers[index].Unobserve(ctx)
				observers = append(observers[:index], observers[index+1:]...)

			case 6: // write a var
				if len(vars) == 0 {
					continue
				}
				vars[int(arg)%len(vars)].Set(int(arg))

			case 7: // stabilize
				if err := g.Stabilize(ctx); err != nil {
					continue
				}

			case 8: // a node that panics, to cover the recovery path's effect on the heap
				if len(nodes) == 0 || len(nodes) >= maxNodes {
					continue
				}
				// recovering from a panic makes the node stale and returns it to the
				// recompute heap, so a panicking node interacts with teardown and with
				// whatever else is queued -- which is the part worth fuzzing rather than
				// the panic itself
				nodes = append(nodes, Map(g, pickNode(arg), func(x int) int {
					if x%3 == 0 {
						panic("a fuzzed node panicked")
					}
					return x + 1
				}))
			}

			if err := ExpertGraph(g).CheckInvariants(); err != nil {
				t.Fatalf("op %d (kind %d, arg %d) left the graph inconsistent: %v", ops, op, arg, err)
			}
		}

		// whatever shape the program built, releasing it must leave nothing behind
		for _, o := range observers {
			o.Unobserve(ctx)
		}
		if err := g.Stabilize(ctx); err != nil {
			t.Fatalf("stabilizing after unobserving everything: %v", err)
		}
		if err := ExpertGraph(g).CheckInvariants(); err != nil {
			t.Fatalf("inconsistent after unobserving everything: %v", err)
		}
		if g.numNodes != 0 {
			t.Fatalf("unobserving everything left %d nodes in the graph", g.numNodes)
		}
		if got := len(g.nodes); got != 0 {
			t.Fatalf("unobserving everything left %d nodes in the graph's node list", got)
		}
		if got := g.recomputeHeap.len(); got != 0 {
			t.Fatalf("unobserving everything left %d nodes in the recompute heap", got)
		}
		if got := len(g.observers); got != 0 {
			t.Fatalf("unobserving everything left %d observers", got)
		}
		// the leak that motivated this: an edge kept on a node that outlives the subgraph
		for _, v := range vars {
			if got := len(v.Node().parents); got != 0 {
				t.Fatalf("a var kept %d dependents after everything was unobserved", got)
			}
		}
	})
}
