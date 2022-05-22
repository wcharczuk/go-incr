package incr

import (
	"context"
	"sync/atomic"
)

// Stabilize kicks off the full stabilization pass given initial nodes
// representing graphs.
//
// The nodes do not need to be any specific type of node in the graph
// as the full graph will be initialized on the first call to stabilize for that graph.
func Stabilize(ctx context.Context, n INode) error {
	if n.Node().g == nil {
		tracePrintf(ctx, "stabilize; initializing graph rooted at: %v", n)
		Initialize(ctx, n)
	}
	if err := stabilize(ctx, n.Node().g); err != nil {
		return err
	}
	return nil
}

func stabilize(ctx context.Context, g *graph) (err error) {
	if err = ensureNotStabilizing(ctx, g); err != nil {
		return
	}
	stabilizeStart(ctx, g)
	defer stabilizeEnd(ctx, g)
	var n INode
	for len(g.recomputeHeap.lookup) > 0 {
		n = g.recomputeHeap.RemoveMin()
		if err = n.Node().recompute(ctx); err != nil {
			return err
		}
	}
	return nil
}

func ensureNotStabilizing(ctx context.Context, g *graph) error {
	if atomic.LoadInt32(&g.status) != StatusNotStabilizing {
		tracePrintf(ctx, "stabilize; already stabilizing, cannot continue")
		return ErrAlreadyStabilizing
	}
	return nil
}

func stabilizeStart(ctx context.Context, g *graph) {
	atomic.StoreInt32(&g.status, StatusStabilizing)
	tracePrintf(ctx, "stabilize[%d]; stabilization starting", g.stabilizationNum)
}

func stabilizeEnd(ctx context.Context, g *graph) {
	defer func() {
		atomic.StoreInt32(&g.status, StatusNotStabilizing)
	}()
	tracePrintf(ctx, "stabilize[%d]; stabilization complete", g.stabilizationNum)
	g.stabilizationNum++
	var n INode
	for g.setDuringStabilization.len > 0 {
		_, n, _ = g.setDuringStabilization.Pop()
		_ = n.Node().maybeStabilize(ctx)
		SetStale(n)
	}
	atomic.StoreInt32(&g.status, StatusRunningUpdateHandlers)
	var updateHandlers []func(context.Context)
	for g.handleAfterStabilization.len > 0 {
		_, updateHandlers, _ = g.handleAfterStabilization.Pop()
		for _, uh := range updateHandlers {
			uh(ctx)
		}
	}
	return
}
