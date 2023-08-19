package incr

import (
	"context"
)

// Stabilize kicks off the full stabilization pass given initial nodes
// representing graphs.
//
// The nodes do not need to be any specific type of node in the graph
// as the full graph will be initialized on the first call to stabilize for that graph.
func (graph *Graph) Stabilize(ctx context.Context) error {
	if err := graph.stabilize(ctx); err != nil {
		return err
	}
	return nil
}

func (graph *Graph) stabilize(ctx context.Context) (err error) {
	if err = graph.ensureNotStabilizing(ctx); err != nil {
		return
	}
	graph.stabilizeStart(ctx)
	defer graph.stabilizeEnd(ctx)
	var n INode
	for len(graph.recomputeHeap.lookup) > 0 {
		n = graph.recomputeHeap.RemoveMin()
		if err = graph.recompute(ctx, n); err != nil {
			return err
		}
	}
	return nil
}
