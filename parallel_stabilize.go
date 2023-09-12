package incr

import (
	"context"
	"fmt"
	"runtime"
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
	graph.stabilizeStart(ctx)
	defer graph.stabilizeEnd(ctx)
	return graph.parallelStabilize(ctx)
}

func (graph *Graph) parallelStabilize(ctx context.Context) error {
	if graph.recomputeHeap.Len() == 0 {
		return nil
	}
	workerPool := new(parallelBatch)
	workerPool.SetLimit(runtime.NumCPU())
	var minHeightBlock []INode
	var err error

	// we have to do this _always_
	var immediateRecompute []INode
	defer func() {
		graph.recomputeHeap.Add(immediateRecompute...)
	}()

	for graph.recomputeHeap.Len() > 0 {
		minHeightBlock = graph.recomputeHeap.RemoveMinHeight()
		if len(minHeightBlock) == 0 {
			return fmt.Errorf("parallel stabilize[%d]; recompute heap has remaining items but min height block is empty, aborting", graph.stabilizationNum)
		}
		for _, n := range minHeightBlock {
			workerPool.Go(graph.parallelRecomputeNode(ctx, n))
			if n.Node().always {
				immediateRecompute = append(immediateRecompute, n)
			}
		}
		if err = workerPool.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (graph *Graph) parallelRecomputeNode(ctx context.Context, n INode) func() error {
	return func() error {
		return graph.recompute(ctx, n)
	}
}
