package incr

import (
	"context"
	"fmt"
	"sync"
)

// ParallelStabilize stabilizes a graph in parallel.
//
// This is done similarly to [Graph.Stabilize], in that nodes are stabilized
// starting with the minimum height in the recompute heap working upwards, but
// unlike the serial processing that [Graph.Stabilize] does, [Graph.ParallelStabilize] will
// process a height "block" all at once concurrently.
//
// Because of the concurrent nature of the block processing, [Graph.ParallelStabilize] is
// considerably slower to process nodes, specifically because locks have to be acquired and shared
// state managed carefully.
//
// You should only reach for [Graph.ParallelStabilize] if you have very long running node recomputations
// that would benefit from processing in parallel, e.g. if you have nodes that are I/O bound or CPU intensive.
//
// Canceling ctx stops the pass at the next height block boundary and returns the
// context's cause; nodes not yet recomputed stay in the recompute heap.
func (graph *Graph) ParallelStabilize(ctx context.Context) (err error) {
	if graph.deterministic {
		err = fmt.Errorf("incr; cannot parallel stabilize if graph is deterministic")
		return
	}
	if err = graph.ensureNotStabilizing(ctx); err != nil {
		return
	}
	ctx = graph.stabilizeStart(ctx)
	defer func() {
		graph.stabilizeEnd(ctx, err)
	}()
	err = graph.parallelStabilize(ctx)
	return
}

func (graph *Graph) parallelStabilize(ctx context.Context) (err error) {
	if graph.recomputeHeap.len() == 0 {
		return
	}

	var immediateRecompute []INode
	var immediateRecomputeMu sync.Mutex
	// Each worker holds its own guard, which is the only place a panic in a worker
	// goroutine can be caught: a recover in the caller never sees it, so without this a
	// panicking node ends the process. The node is a local, so nothing is shared.
	parallelRecomputeNode := func(ctx context.Context, n INode) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = graph.recomputePanicked(ctx, n, r)
			}
		}()
		err = graph.recompute(ctx, n, true)
		if n.Node().always {
			immediateRecomputeMu.Lock()
			immediateRecompute = append(immediateRecompute, n)
			immediateRecomputeMu.Unlock()
		}
		return
	}

	// Cancellation is checked once per height block rather than per node: the batch is
	// where a pass can be interrupted without leaving a block half applied, and parallel
	// stabilization is for nodes expensive enough that block granularity is fine.
	done := ctx.Done()

	var iter recomputeHeapListIter
	for graph.recomputeHeap.len() > 0 {
		if err = contextCanceled(ctx, done); err != nil {
			break
		}
		graph.recomputeHeap.setIterToMinHeight(&iter)
		err = parallelBatch(ctx, parallelRecomputeNode, iter.Next, graph.parallelism)
		if err != nil {
			break
		}
	}
	if err != nil {
		if graph.clearRecomputeHeapOnError {
			aborted := graph.recomputeHeap.clear()
			for _, node := range aborted {
				for _, ah := range node.Node().abortedHandlers() {
					ah(ctx, err)
				}
			}
		}
	}
	if len(immediateRecompute) > 0 {
		graph.recomputeHeap.mu.Lock()
		for _, n := range immediateRecompute {
			if n.Node().heightInRecomputeHeap == HeightUnset {
				graph.recomputeHeap.addNodeUnsafe(n)
			}
		}
		graph.recomputeHeap.mu.Unlock()
	}
	return
}
