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

	var immediateRecompute []INode
	// we have to do this _always_
	defer func() {
		graph.recomputeHeap.Add(immediateRecompute...)
	}()

	for len(graph.recomputeHeap.lookup) > 0 {
		n = graph.recomputeHeap.RemoveMin()
		if err = graph.recompute(ctx, n); err != nil {
			TraceErrorf(ctx, "stabilize[%d]; node recompute error %v: %+v", graph.stabilizationNum, n, err)
			return err
		}
		if n.Node().always {
			TracePrintf(ctx, "stabilize[%d]; adding always node to immediate recompute list %v", graph.stabilizationNum, n)
			immediateRecompute = append(immediateRecompute, n)
		}
	}
	return nil
}
