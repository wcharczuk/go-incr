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
		if seenGraphs.has(gn.Node().gs.id) {
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
	if gnn.gs.status != StatusNotStabilizing {
		tracePrintf(ctx, "stabilize; already stabilizing, cannot continue")
		return fmt.Errorf("stabilize; already stabilizing, cannot continue")
	}
	gnn.gs.mu.Lock()
	defer gnn.gs.mu.Unlock()
	defer func() {
		tracePrintf(ctx, "stabilize[%d]; stabilization complete", gnn.gs.stabilizationNum)
		gnn.gs.stabilizationNum++
		gnn.gs.status = StatusNotStabilizing
	}()
	gnn.gs.status = StatusStabilizing
	tracePrintf(ctx, "stabilize[%d]; stabilization starting", gnn.gs.stabilizationNum)
	return recomputeAll(ctx, gnn.gs)
}

func recomputeAll(ctx context.Context, gs *graphState) error {
	var err error
	var n INode
	var nn *Node
	for gs.rh.Len() > 0 {
		n = gs.rh.RemoveMin()
		nn = n.Node()
		if nn.shouldRecompute() {
			if nn.maybeCutoff(ctx) {
				return nil
			}
			nn.changedAt = nn.gs.stabilizationNum
			if err = nn.recompute(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
