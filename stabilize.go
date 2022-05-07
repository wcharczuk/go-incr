package incr

import (
	"context"
)

// Stabilize kicks off the full stabilization pass given an initial node
// connected to the graph.
//
// The node does not need to be an input, outpoot, root or leaf node
// in the graph, the full graph will be discovered, and initialized
// on the first call to stabilize, and evaluated subsequently each pass.
func Stabilize(ctx context.Context, gn GraphNode) error {
	gnn := gn.Node()
	if shouldInitialize(gnn) {
		tracePrintf(ctx, "stabilize; initializing graph rooted at: %v", gn)
		Initialize(ctx, gn)
	}
	defer func() {
		gnn.gs.sn++
		gnn.gs.s = StatusNotStabilizing
		tracePrintf(ctx, "stabilize; stabilization %d complete", gnn.gs.sn)
	}()
	gnn.gs.s = StatusStabilizing
	tracePrintf(ctx, "stabilize; stabilization %d starting", gnn.gs.sn)
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
	var n GraphNode
	var nn *Node
	for gs.rh.len > 0 {
		n = gs.rh.removeMin()
		nn = n.Node()
		if nn.stale(ctx) {
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
