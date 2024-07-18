package incr

import (
	"context"
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
func (graph *Graph) ParallelStabilize(ctx context.Context) (err error) {
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
	parallelRecomputeNode := func(ctx context.Context, n INode) (err error) {
		err = graph.recompute(ctx, n, true)
		if n.Node().always {
			immediateRecomputeMu.Lock()
			immediateRecompute = append(immediateRecompute, n)
			immediateRecomputeMu.Unlock()
		}
		return
	}

	var iter recomputeHeapListIter
	for graph.recomputeHeap.len() > 0 {
		graph.recomputeHeap.setIterToMinHeight(&iter)
		err = parallelBatch[INode](ctx, parallelRecomputeNode, iter.Next, graph.parallelism)
		if err != nil {
			break
		}
	}
	if err != nil {
		// clear if there is an error!
		aborted := graph.recomputeHeap.clear()
		for _, node := range aborted {
			for _, ah := range node.Node().onAbortedHandlers {
				ah(ctx, err)
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
