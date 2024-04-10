package incr

import (
	"context"
)

// Stabilize kicks off the stabilization for nodes that have been observed by the graph's scope.
//
// The general process of stabilization is to scan the recompute heap up from the minimum height
// block, processing each node in the block until there are no more nodes.
//
// Stabilizing a node can add more nodes to the recompute heap, creating more work as the stabilization
// progresses, until finally no more nodes are left to process.
//
// The [Stabailize] stabilization process is serial, that is each node is recomputed in sequence one
// after the other.
//
// This can be extremely fast in practice because it lets us makes some assumptions about what
// can change in during each node's stabilization, specifically we can assume that [Bind] nodes
// evaluate serially and we can adjust recompute heights accordingly, and as a result we don't need
// to worry about shared resource contention and can skip acquiring locks.
//
// If during the stabilization pass a node's stabilize function returns an error, the recomputation pass
// is stopped and the error is returned.
func (graph *Graph) Stabilize(ctx context.Context) (err error) {
	if err = graph.ensureNotStabilizing(ctx); err != nil {
		return
	}
	ctx = graph.stabilizeStart(ctx)
	defer func() {
		graph.stabilizeEnd(ctx, err)
	}()

	var immediateRecompute []INode
	var next INode
	for graph.recomputeHeap.numItems > 0 {
		next, _ = graph.recomputeHeap.removeMinUnsafe()
		err = graph.recompute(ctx, next, false /*parallel*/)
		if next.Node().always {
			immediateRecompute = append(immediateRecompute, next)
		}
		if err != nil {
			break
		}
	}
	if len(immediateRecompute) > 0 {
		for _, n := range immediateRecompute {
			graph.recomputeHeap.addIfNotPresent(n)
		}
	}
	return
}
