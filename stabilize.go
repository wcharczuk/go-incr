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
	graph.stabilizeStart(ctx)
	defer func() {
		graph.stabilizeEnd(ctx, err)
	}()
	var n INode

	var immediateRecompute []INode
	for len(graph.recomputeHeap.lookup) > 0 {
		n = graph.recomputeHeap.RemoveMin()
		if err = graph.recompute(ctx, n); err != nil {
			break
		}
		if n.Node().always {
			immediateRecompute = append(immediateRecompute, n)
		}
	}

	graph.recomputeHeap.Add(immediateRecompute...)
	return
}
