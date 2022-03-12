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

	discoverySeen := make(Set[NodeID])
	recompute := &Heap[Stabilizer]{
		LessFn: nodeHeightLess,
	}

	var generation, latestGeneration Generation
	var n Stabilizer
	var nn *Node
	var id NodeID
	var ok bool
	for {
		n, ok = discovery.Pop()
		if !ok {
			break
		}
		nn = n.Node()
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
			tracePrintf(ctx, "stabilize; marking to recompute; %T %v", n, id)
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

	recomputeSeen := make(Set[NodeID])

	var err error
	var cid NodeID
	for {
		n, ok = recompute.Pop()
		if !ok {
			break
		}
		nn = n.Node()
		tracePrintf(ctx, "stabilize; recomputing; %T %v", n, id)
		if err = n.Stabilize(ctx, latestGeneration); err != nil {
			return err
		}
		if nn.recomputedAt < nn.changedAt {
			for _, c := range nn.children {
				cid = c.Node().id
				if recomputeSeen.Has(cid) {
					continue
				}
				recompute.Push(c)
				recomputeSeen.Add(cid)
			}
		}
		// these are down here to not foul up the if statement above
		// and it should always be applied
		if !nn.initialized {
			nn.initialized = true
		}
		nn.recomputedAt = latestGeneration
	}
	return nil
}

func shouldRecompute(s Stabilizer, latestGeneration Generation) bool {
	return !s.Node().initialized || s.Node().changedAt > latestGeneration
}

func nodeHeightLess(a, b Stabilizer) bool {
	return a.Node().height < b.Node().height
}
