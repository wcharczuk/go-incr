package incr

import "context"

// initialize starts the initialization cycle.
//
// it creates the graph state for the graph, discovers and initialized
// all nodes, and then establishes the recompute heap based on node heights.
func Initialize(ctx context.Context, nodes ...INode) {
	for _, n := range nodes {
		if !shouldInitialize(n.Node()) {
			continue
		}
		discoverAllNodes(ctx, newGraphState(), n)
	}
}

// shouldInitialize returns if the graph is uninitialized
//
// specifically if it needs to have the first pass of initialization
// performed on it, setting up the graph state, the recompute heap,
// and other node metadata items.
func shouldInitialize(n *Node) bool {
	return n.gs == nil
}

func discoverAllNodes(ctx context.Context, gs *graphState, gn INode) {
	discoverNode(ctx, gs, gn)
	gnn := gn.Node()
	for _, c := range gnn.children {
		if !shouldInitialize(c.Node()) {
			continue
		}
		discoverAllNodes(ctx, gs, c)
	}
	for _, p := range gnn.parents {
		if !shouldInitialize(p.Node()) {
			continue
		}
		discoverAllNodes(ctx, gs, p)
	}
}

func discoverNode(ctx context.Context, gs *graphState, gn INode) {
	gnn := gn.Node()
	gnn.gs = gs
	gnn.detectCutoff(gn)
	gnn.detectStabilize(gn)
	gnn.height = gnn.calculateHeight()
	gs.rh.Add(gn)
	return
}

// undiscoverAllNodes removes a node and all its parents
// from a given graph.
//
// NOTE: you _must_ unlink it first or you'll just blow away the whole graph.
func undiscoverAllNodes(ctx context.Context, gs *graphState, gn INode) {
	undiscoverNode(ctx, gs, gn)
	gnn := gn.Node()
	for _, c := range gnn.children {
		if shouldInitialize(c.Node()) {
			continue
		}
		undiscoverAllNodes(ctx, gs, c)
	}
	for _, p := range gnn.parents {
		if shouldInitialize(p.Node()) {
			continue
		}
		undiscoverAllNodes(ctx, gs, p)
	}
}

func undiscoverNode(ctx context.Context, gs *graphState, gn INode) {
	gnn := gn.Node()
	gnn.gs = nil
	gs.rh.Remove(gn)
}
