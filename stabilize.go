package incr

import (
	"context"
)

// Stabilize traverses the DAG in topological order starting at variables that changed
// since the last stabilization and recomputing their dependents. This is done by using
// a "recompute heap" to visit the nodes in non-decreasing order of "height", which is a
// over-approximation of the longest path from a variable to that node.  To ensure that
// each node is computed at most once and that its children are stabilized before it is
// computed, nodes satisfy the property that if there is an edge from n1 to n2, then the
// height of n1 is less than the height of n2.
//
// Stabilize repeats the following steps until the heap becomes empty:
//    1. remove from the recompute heap a node with the smallest height
//    2. recompute that node
//    3. if the node's value changes, then add its parents to the heap.
//
// The definition of "changes" in step (3) is configurable by user code. By default, a
// node is considered to change if its new value is not equal to the previous
// value. One can use Cutoff[A] on a node to change its cutoff function, e.g. for
// float64 one could cutoff propagation if the old value and new value are closer than
// some epsilon.
//
// Stabilization uses a heap implemented with an array whose length is the max height, so
// for good performance, the height of nodes must be small.  There is an upper bound on
// the height of nodes, [max_height_allowed], which defaults to 128. An attempt to
// create a node with larger height will return an error.
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

	// discover all nodes
	var n Stabilizer
	for discovery.Len() > 0 {
		n, _ = discovery.Pop()
		if n.Stale() {
			if !recomputeSeen.Has(n.getNode().id) {
				recomputeSeen.Add(n.getNode().id)
				recompute.Push(n)
			}
		}
		for _, p := range n.getNode().parents {
			discovery.Push(p)
		}
	}

	tracePrintf(ctx, "stabilize recompute queue: %d nodes", recompute.Len())

	var err error
	for recompute.Len() > 0 {
		n, _ = recompute.Pop()
		if err = n.Stabilize(ctx); err != nil {
			return err
		}
	}
	return nil
}

func stabilizeFindRecomputeChildren(recompute *Heap[Stabilizer], recomputeSeen map[nodeID]struct{}, n Stabilizer) {
	recompute.Push(n)
	for _, p := range n.getNode().children {
		stabilizeFindRecomputeChildren(recompute, recomputeSeen, p)
	}
}

func nodeHeightLess(a, b Stabilizer) bool {
	return a.getNode().height < b.getNode().height
}
