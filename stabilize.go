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
		tracePrintf(ctx, "stabilize; initializing graph rooted at: %s", gn.String())
		if err := Initialize(ctx, gn); err != nil {
			return err
		}
	}
	defer func() {
		gnn.gs.sn++
		gnn.gs.s = StatusNotStabilizing
	}()
	gnn.gs.s = StatusStabilizing
	tracePrintf(ctx, "stabilize; beginning stabilization %d", gnn.gs.sn)
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
	tracePrintf(ctx, "stabilize; recompute; %d node in heap", gs.rh.len)
	for gs.rh.len > 0 {
		n = gs.rh.removeMin()
		nn = n.Node()
		if nn.stale(ctx) {
			if nn.maybeCutoff(ctx) {
				tracePrintf(ctx, "stabilize; recompute; skipping %s, fails cutoff", n.String())
				continue
			}
			nn.changedAt = gs.sn
			tracePrintf(ctx, "stabilize; recompute; stabilizing %s", n.String())
			if err = nn.recompute(ctx); err != nil {
				return err
			}
		} else {
			tracePrintf(ctx, "stabilize; recompute; skipping %s, not stale", n.String())
		}
	}
	return nil
}
