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
	defer func() {
		graph.stabilizeEnd(ctx, err)
	}()
	err = graph.parallelStabilize(ctx)
	return
}

func (graph *Graph) parallelStabilize(ctx context.Context) (err error) {
	if graph.recomputeHeap.Len() == 0 {
		return
	}
	workerPool := new(parallelBatch)
	workerPool.SetLimit(runtime.NumCPU())

	// we have to do this _always_
	var immediateRecompute []INode

	var minHeightBlock []INode
	for graph.recomputeHeap.Len() > 0 {
		minHeightBlock = graph.recomputeHeap.RemoveMinHeight()
		if len(minHeightBlock) == 0 {
			// NOTE(wc): this is a """should be impossible""" edge case that can come up if the user is managing nodes invasively,
			// specifically we can get into situations where the node height changes and is not reflected in the heap height lists accordingly.
			//
			// If this check is not here, and this condition is present, the stabilization _will never finish_ because there are items
			// left in the heap lookup but they are not in the correct height "block", leading to an infinite loop.
			err = fmt.Errorf("parallel stabilize[%d]; recompute heap has remaining items but min height block is empty, aborting", graph.stabilizationNum)
			break
		}
		for _, n := range minHeightBlock {
			workerPool.Go(graph.parallelRecomputeNode(ctx, n))
			if n.Node().always {
				immediateRecompute = append(immediateRecompute, n)
			}
		}
		if err = workerPool.Wait(); err != nil {
			break
		}
	}
	graph.recomputeHeap.Add(immediateRecompute...)
	return
}

func (graph *Graph) parallelRecomputeNode(ctx context.Context, n INode) func() error {
	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		err = graph.recompute(ctx, n)
		return
	}
}
