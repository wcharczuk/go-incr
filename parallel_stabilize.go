package incr

import (
	"context"
)

// ParallelStabilize stabilizes graphs in parallel as entered
// through representative nodes.
//
// For each input node, if the node is not attached to a graph, the full
// graph is discovered from the relationships on that given input node.
//
// If multiple nodes are supplied that are actually connected, initialization
// will skip the already connected (and as a result, initialized) graph.
//
// ParallelStabilize differs from Stabilize in that it reads the current
// recompute heap in pseudo-height chunks, processing each pseudo-height in
// parallel before moving on to the next, smallest height chunk.
//
// Each parallel recompute cycle may produce new nodes to process, and as a result
// parallel stabilization can move up and down in height before fully recomputing
// the graph.
func (graph *Graph) ParallelStabilize(ctx context.Context) (err error) {
	if err = graph.ensureNotStabilizing(ctx); err != nil {
		return
	}
	ctx = graph.stabilizeStart(ctx)
	defer func() {
		graph.stabilizeEnd(ctx, err, true /*parallel*/)
	}()
	err = graph.parallelStabilize(ctx)
	return
}

func (graph *Graph) parallelStabilize(ctx context.Context) (err error) {
	if graph.recomputeHeap.len() == 0 {
		return
	}

	var immediateRecompute []INode
	var minHeightBlock []INode
	for graph.recomputeHeap.len() > 0 {
		minHeightBlock = graph.recomputeHeap.removeMinHeight()
		for _, n := range minHeightBlock {
			graph.workerPool.Go(graph.parallelRecomputeNode(ctx, n))
			if n.Node().always {
				immediateRecompute = append(immediateRecompute, n)
			}
		}
		if err = graph.workerPool.Wait(); err != nil {
			break
		}
	}
	if len(immediateRecompute) > 0 {
		graph.recomputeHeap.add(immediateRecompute...)
	}
	return
}

func (graph *Graph) parallelRecomputeNode(ctx context.Context, n INode) func() error {
	return func() (err error) {
		err = graph.recompute(ctx, n, true)
		return
	}
}
