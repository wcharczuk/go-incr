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
// value. One can use `Cutoff[A]` on a node to change its cutoff function, e.g. for
// `float64` one could cutoff propagation if the old value and new value are closer than
// some epsilon.
//
// Stabilization uses a heap implemented with an array whose length is the max height, so
// for good performance, the height of nodes must be small.  There is an upper bound on
// the height of nodes, `MaxHeightAllowed`, which defaults to 128. An attempt to
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

	// discover stale nodes; these are typically variables
	// and bind nodes that have to recompute each pass
	var n Stabilizer
	var id nodeID
	for discovery.Len() > 0 {
		n, _ = discovery.Pop()
		id = n.getNode().id
		if isDiscoveryStale(n) {
			if !recomputeSeen.Has(id) {
				recomputeSeen.Add(id)
				recompute.Push(n)
			}
		}
		for _, p := range n.getNode().parents {
			discovery.Push(p)
		}
	}

	tracePrintf(ctx, "stabilize; computation has %d stale or dynamic nodes", recompute.Len())

	var err error
	for recompute.Len() > 0 {
		n, _ = recompute.Pop()
		if err = n.Stabilize(ctx); err != nil {
			return err
		}

		// given we just stabilized the variable or bind
		// find all the children of these nodes and mark them
		// to be recomputed also
		stabilizeRecomputeFindChildren(ctx, recompute, recomputeSeen, n)
	}
	return nil
}

func stabilizeRecomputeFindChildren(ctx context.Context, recompute *Heap[Stabilizer], recomputeSeen Set[nodeID], n Stabilizer) {
	for _, c := range n.getNode().children {
		cid := c.getNode().id
		if recomputeSeen.Has(cid) {
			continue
		}
		if !isRecomputeStale(c) {
			continue
		}
		recompute.Push(c)
		recomputeSeen.Add(cid)
		stabilizeRecomputeFindChildren(ctx, recompute, recomputeSeen, c)
	}
}

// isDiscoveryStale is used to determine which nodes
// are initially stale; typically these are variables that have
// been modified since the last stabilization, or dynamic nodes.
//
// it defaults to `false` unless the node specifically
// implements `Staler` and returns true.
func isDiscoveryStale(n Stabilizer) bool {
	if typed, ok := n.(Staler); ok {
		return typed.Stale()
	}
	return false
}

// isRecomputeStale is used to determine which children
// of a stale input or dynamic should be stabilized.
//
// it defaults to `true` unless the node specifically
// implements `Staler` and returns false.
func isRecomputeStale(n Stabilizer) bool {
	if typed, ok := n.(Staler); ok {
		return typed.Stale()
	}
	return true
}

func nodeHeightLess(a, b Stabilizer) bool {
	return a.getNode().height < b.getNode().height
}
