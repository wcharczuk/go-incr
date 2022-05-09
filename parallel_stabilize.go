package incr

import (
	"context"
	"fmt"
	"sync"
)

// ParallelStabilize kicks off the full stabilization pass given initial nodes
// representing graphs
//
// The nodes do not need to be any specific type of node in the graph
// as the full graph will be initialized on the first call to stabilize for that graph.
func ParallelStabilize(ctx context.Context, nodes ...INode) error {
	for _, gn := range nodes {
		if err := parallelStabilizeNode(ctx, gn); err != nil {
			return err
		}
	}
	return nil
}

func parallelStabilizeNode(ctx context.Context, gn INode) error {
	gnn := gn.Node()
	if shouldInitialize(gnn) {
		return fmt.Errorf("cannot parallel stabilize; must initialize first")
	}
	defer func() {
		tracePrintf(ctx, "parallel stabilize; stabilization %s.%d complete", gnn.gs.id.Short(), gnn.gs.sn)
		gnn.gs.sn++
		gnn.gs.s = StatusNotStabilizing
	}()
	gnn.gs.s = StatusStabilizing
	tracePrintf(ctx, "parallel stabilize; stabilization %s.%d starting", gnn.gs.id.Short(), gnn.gs.sn)
	return parallelRecomputeAll(ctx, gnn.gs)
}

func parallelRecomputeAll(ctx context.Context, gs *graphState) error {
	processing := sync.WaitGroup{}
	wp := newWorkerPool(func(ictx context.Context, nn *Node) error {
		nn.changedAt = gs.sn
		return nn.recompute(ictx)
	}, func(_ context.Context, _ *worker[*Node]) error {
		processing.Done()
		return nil
	})
	if err := wp.Start(ctx); err != nil {
		return err
	}
	defer func() { _ = wp.Shutdown(ctx) }()

process:
	for gs.rh.Len() > 0 {
		n := gs.rh.RemoveMin()
		nn := n.Node()
		recompute, err := nn.shouldRecompute(ctx)
		if err != nil {
			return err
		}
		if recompute {
			if nn.maybeCutoff(ctx) {
				continue
			}
			processing.Add(1)
			wp.Push(ctx, nn)
		}
	}
	processing.Wait()

	if gs.rh.Len() > 0 {
		// goto considered harmful ... kind of.
		goto process
	}
	return nil
}
