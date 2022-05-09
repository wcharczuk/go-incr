package incr

import (
	"context"
)

// Stabilize kicks off the full stabilization pass given initial nodes
// representing graphs.
//
// The nodes do not need to be any specific type of node in the graph
// as the full graph will be initialized on the first call to stabilize for that graph.
func Stabilize(ctx context.Context, nodes ...INode) error {
	for _, gn := range nodes {
		if err := stabilizeNode(ctx, gn); err != nil {
			return err
		}
	}
	return nil
}

func stabilizeNode(ctx context.Context, gn INode) error {
	gnn := gn.Node()
	if shouldInitialize(gnn) {
		tracePrintf(ctx, "stabilize; initializing graph rooted at: %v", gn)
		Initialize(ctx, gn)
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

// shouldInitialize returns if the graph is uninitialized
//
// specifically if it needs to have the first pass of initialization
// performed on it, setting up the graph state, the recompute heap,
// and other node metadata items.
func shouldInitialize(n *Node) bool {
	return n.gs == nil
}

func recomputeAll(ctx context.Context, gs *graphState) error {
	var err error
	var n INode
	var nn *Node
	var recompute bool
	for gs.rh.Len() > 0 {
		n = gs.rh.RemoveMin()
		nn = n.Node()
		recompute, err = nn.shouldRecompute(ctx)
		if err != nil {
			return err
		}
		if recompute {
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
