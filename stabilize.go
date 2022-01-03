package incr

import (
	"context"
)

// Stabilize recomputes a graph of computations rooted at a given list of outputs.
//
// It does this by traversing the DAG, working "up" from the output nodes, by iterating
// through the parents, and then the parents of those nodes parents etc., until it
// discovers all the Variable nodes in the graph, adding them to a recomputation heap.
//
// Once all the variables are discovered, recomputation proceeds as
// follows (until the heap is empty):
// - a minimum-height node is popped off the recompute heap
// - it is recomputed, if the recomputation returns an error it is returned
//   and recomputation stops
// - if the value of the node has changed, all of that nodes children are
//   added to the recomputation heap
//
// Stabilize is safe to run multiple times if the variables haven't changed.
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

func isDiscoveryRecompute(s Stabilizer) bool {
	return s.getNode().isVariable
}

func nodeHeightLess(a, b Stabilizer) bool {
	return a.getNode().height < b.getNode().height
}
