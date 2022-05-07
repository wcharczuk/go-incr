package incr

import "context"

// initialize starts the initialization cycle.
//
// it creates the graph state for the graph, discovers and initialized
// all nodes, and then establishes the recompute heap based on node heights.
func Initialize(ctx context.Context, gn GraphNode) (err error) {
	gs := newGraphState()
	err = discoverAllNodes(ctx, gs, gn)
	if err != nil {
		return err
	}
	return
}

func discoverAllNodes(ctx context.Context, gs *graphState, gn GraphNode) error {
	if err := discoverNode(ctx, gs, gn); err != nil {
		return err
	}
	gnn := gn.Node()
	for _, c := range gnn.children {
		if !shouldInitialize(c.Node()) {
			continue
		}
		if err := discoverAllNodes(ctx, gs, c); err != nil {
			return err
		}
	}
	for _, p := range gnn.parents {
		if !shouldInitialize(p.Node()) {
			continue
		}
		if err := discoverAllNodes(ctx, gs, p); err != nil {
			return err
		}
	}
	return nil
}

func discoverNode(ctx context.Context, gs *graphState, s GraphNode) (err error) {
	sn := s.Node()
	sn.gs = gs
	sn.detectCutoff(s)
	sn.detectStabilizer(s)
	sn.height = sn.calculateHeight()
	gs.rh.add(s)
	return
}
