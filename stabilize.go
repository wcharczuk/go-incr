package incr

import (
	"context"
	"fmt"
)

// Stabilize kicks off the full stabilization pass given initial nodes
// representing graphs.
//
// The nodes do not need to be any specific type of node in the graph
// as the full graph will be initialized on the first call to stabilize for that graph.
func Stabilize(ctx context.Context, nodes ...INode) error {
	seenGraphs := make(set[Identifier])
	for _, gn := range nodes {
		if shouldInitialize(gn.Node()) {
			tracePrintf(ctx, "stabilize; initializing graph rooted at: %v", gn)
			Initialize(ctx, gn)
		}
		if seenGraphs.has(gn.Node().g.id) {
			continue
		}
		if err := stabilizeNodeGraph(ctx, gn); err != nil {
			return err
		}
	}
	return nil
}

func stabilizeNodeGraph(ctx context.Context, gn INode) error {
	gnn := gn.Node()
	if gnn.g.status != StatusNotStabilizing {
		tracePrintf(ctx, "stabilize; already stabilizing, cannot continue")
		return fmt.Errorf("stabilize; already stabilizing, cannot continue")
	}
	gnn.g.mu.Lock()
	defer gnn.g.mu.Unlock()
	defer func() {
		tracePrintf(ctx, "stabilize[%d]; stabilization complete", gnn.g.stabilizationNum)
		gnn.g.stabilizationNum++
		gnn.g.status = StatusNotStabilizing
	}()
	gnn.g.status = StatusStabilizing
	tracePrintf(ctx, "stabilize[%d]; stabilization starting", gnn.g.stabilizationNum)
	return stabilize(ctx, gnn.g)
}

func stabilize(ctx context.Context, g *graph) error {
	var err error
	var n INode
	for len(g.rh.lookup) > 0 {
		n = g.rh.RemoveMin()
		if err = n.Node().recompute(ctx); err != nil {
			return err
		}
	}
	return nil
}
