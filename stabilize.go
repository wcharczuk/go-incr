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
		LessFn: nodeHeightLess,
	}
	discovery.Init()

	discoverySeen := make(Set[nodeID])
	recompute := &Heap[Stabilizer]{
		LessFn: nodeHeightLess,
	}

	var generation, latestGeneration generation
	var n Stabilizer
	var nn *node
	var id nodeID
	var ok bool
	for {
		n, ok = discovery.Pop()
		if !ok {
			break
		}
		nn = n.getNode()
		id = nn.id
		generation = nn.recomputedAt
		if generation > latestGeneration {
			tracePrintf(ctx, "stabilize; updating latest generation; %d", generation)
			latestGeneration = generation
		}

		if discoverySeen.Has(id) {
			continue
		}
		if shouldRecompute(n, latestGeneration) {
			tracePrintf(ctx, "stabilize; recomputing; %T %v", n, id)
			recompute.Push(n)
		} else {
			tracePrintf(ctx, "stabilize; skipping; %T %v", n, id)
		}

		discoverySeen.Add(id)
		for _, p := range nn.parents {
			discovery.Push(p)
		}
	}

	latestGeneration = latestGeneration + 1
	tracePrintf(ctx, "stabilize; computation at generation %d has %d stale nodes", latestGeneration, recompute.Len())

	recomputeSeen := make(Set[nodeID])

	var err error
	var before, after any
	var cid nodeID
	for {
		n, ok = recompute.Pop()
		if !ok {
			break
		}

		before = n.getValue()
		nn = n.getNode()
		if err = n.Stabilize(ctx); err != nil {
			return err
		}

		after = n.getValue()
		if before != after || nn.recomputedAt < nn.changedAt {
			for _, c := range nn.children {
				cid = c.getNode().id
				if recomputeSeen.Has(cid) {
					continue
				}
				recompute.Push(c)
				recomputeSeen.Add(cid)
			}
		}
		// this is down here to not foul up the if statement above
		// and it should always be set
		nn.recomputedAt = latestGeneration
	}
	return nil
}

func shouldRecompute(s Stabilizer, latestGeneration generation) bool {
	return !s.getNode().initialized || s.getNode().changedAt > latestGeneration
}

func nodeHeightLess(a, b Stabilizer) bool {
	return a.getNode().height < b.getNode().height
}
