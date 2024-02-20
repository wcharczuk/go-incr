package incr

import (
	"context"
)

// Stabilize kicks off the full stabilization pass given initial nodes
// representing graphs.
//
// The nodes do not need to be any specific type of node in the graph
// as the full graph will be initialized on the first call to stabilize for that graph.
func (graph *Graph) Stabilize(ctx context.Context) (err error) {
	if err = graph.ensureNotStabilizing(ctx); err != nil {
		return
	}
	ctx = graph.stabilizeStart(ctx)
	defer func() {
		graph.stabilizeEnd(ctx, err, false /*parallel*/)
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
		graph.recomputeHeap.add(immediateRecompute...)
	}
	return
}
