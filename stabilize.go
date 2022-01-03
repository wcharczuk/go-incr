package incr

import (
	"context"
)

// Stabilize recomputes a graph of computations rooted at a given list of outputs.
//
// It does this by traversing the DAG, discovering any `Var` or `Bind...` nodes
// and marking them to be recomputed if their value has changed (referred to as being "stale").
//
// Once they're recomputed, any stale "children" of those nodes are also recomputed recursively.
//
// The order of nodes to be recomputed leverages a min-heap ordered by the maximum
// height of a given node from its inner-most parent.
//
// Non-dynamic (Map, Cutoff etc.) nodes are only recomputed if their inputs change.
// If you need to have a node recompute _always_ regardless of if the inputs change,
// you should use a bind-style node.
func Stabilize(ctx context.Context, outputs ...Stabilizer) error {
	discovery := &Heap[Stabilizer]{
		Values: outputs,
		Less:   nodeHeightLess,
	}
	discovery.Init()

	recomputeSeen := make(Set[nodeID])
	recompute := &Heap[Stabilizer]{
		Values: nil,
		Less:   nodeHeightLess,
	}
	recompute.Init()

	// discover stale nodes; these are typically variables
	// and bind nodes that have to recompute each pass
	var n Stabilizer
	var id nodeID
	for discovery.Len() > 0 {
		n, _ = discovery.Pop()
		id = n.getNode().id
		if isDiscoveryRecompute(n) {
			if !recomputeSeen.Has(id) {
				recomputeSeen.Add(id)
				recompute.Push(n)
			}
		}
		for _, p := range n.getNode().parents {
			discovery.Push(p)
		}
	}

	tracePrintf(ctx, "stabilize; computation has %d stale or bind nodes", recompute.Len())

	var err error
	var before, after any
	// we only recompute "stale" nodes, or nodes whos value
	// has changed (in the case of var's), or if they're bind/dynamic nodes.
	for recompute.Len() > 0 {
		n, _ = recompute.Pop()
		before = n.getValue()
		if err = n.Stabilize(ctx); err != nil {
			return err
		}
		after = n.getValue()

		if before != after {
			for _, c := range n.getNode().children {
				cid := c.getNode().id
				if recomputeSeen.Has(cid) {
					continue
				}
				recompute.Push(c)
				recomputeSeen.Add(cid)
			}
		}
	}
	return nil
}

func getValue[A comparable](s Stabilizer) (output A) {
	typed, ok := s.(Incr[A])
	if ok {
		output = typed.Value()
		return
	}
	return
}

func isDiscoveryRecompute(s Stabilizer) bool {
	return s.getNode().isVariable
}

func nodeHeightLess(a, b Stabilizer) bool {
	return a.getNode().height < b.getNode().height
}
