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
		if err := stabilizeNode(ctx, gn); err != nil {
			return err
		}
	}
	return nil
}

func stabilizeNode(ctx context.Context, gn INode) error {
	gnn := gn.Node()
	gnn.gs.mu.Lock()
	defer gnn.gs.mu.Unlock()
	if gnn.gs.s != StatusNotStabilizing {
		tracePrintf(ctx, "stabilize; already stabilizing, cannot continue")
		return fmt.Errorf("stabilize; already stabilizing, cannot continue")
	}
	defer func() {
		tracePrintf(ctx, "stabilize; stabilization %s.%d complete", gnn.gs.id.Short(), gnn.gs.sn)
		gnn.gs.sn++
		gnn.gs.s = StatusNotStabilizing
	}()
	gnn.gs.s = StatusStabilizing
	tracePrintf(ctx, "stabilize; stabilization %s.%d starting", gnn.gs.id.Short(), gnn.gs.sn)
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
				continue
			}
			nn.changedAt = gs.sn
			if err = nn.recompute(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
