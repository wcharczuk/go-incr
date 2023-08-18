package incr

import (
	"context"
	"fmt"
	"runtime"
	"sync"
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
	return graph.parallelRecomputeAll(ctx)
}

func (graph *Graph) parallelRecomputeAll(ctx context.Context) error {
	if graph.recomputeHeap.Len() == 0 {
		return nil
	}
	workerPool := new(parallelBatch)
	workerPool.SetLimit(runtime.NumCPU())
	var minHeight int
	var minHeightBlock []INode
	var err error
	for graph.recomputeHeap.Len() > 0 {
		minHeight = graph.recomputeHeap.minHeight
		minHeightBlock = graph.recomputeHeap.RemoveMinHeight()
		tracePrintf(ctx, "parallel stabilize; recomputing height %d block with %d nodes", minHeight, len(minHeightBlock))
		for _, n := range minHeightBlock {
			workerPool.Go(graph.parallelRecomputeNode(ctx, n.Node()))
		}
		if err = workerPool.Wait(); err != nil {
			return err
		}
	}
	return nil
}

func (graph *Graph) parallelRecomputeNode(ctx context.Context, n *Node) func() error {
	return func() error {
		return graph.recompute(ctx, n)
	}
}

// parallelBatch is a collection of goroutines working on subtasks that are part of
// the same overall task.
//
// A zero Group is valid and does not cancel on error.
type parallelBatch struct {
	wg      sync.WaitGroup
	sem     chan parallelBatchToken
	errOnce sync.Once
	err     error
}

// parallelBatchToken is a token within the parallel batch system
type parallelBatchToken struct{}

func (pb *parallelBatch) done() {
	if pb.sem != nil {
		<-pb.sem
	}
	pb.wg.Done()
}

// Wait blocks until all function calls from the Go method have returned, then
// returns the first non-nil error (if any) from them.
func (pb *parallelBatch) Wait() error {
	pb.wg.Wait()
	return pb.err
}

// Go calls the given function in a new goroutine.
// It blocks until the new goroutine can be added without the number of
// active goroutines in the group exceeding the configured limit.
//
// The first call to return a non-nil error cancels the group; its error will be
// returned by Wait.
func (pb *parallelBatch) Go(f func() error) {
	if pb.sem != nil {
		pb.sem <- parallelBatchToken{}
	}

	pb.wg.Add(1)
	go func() {
		defer pb.done()

		if err := f(); err != nil {
			pb.errOnce.Do(func() {
				pb.err = err
			})
		}
	}()
}

// SetLimit limits the number of active goroutines in this group to at most n.
// A negative value indicates no limit.
//
// Any subsequent call to the Go method will block until it can add an active
// goroutine without exceeding the configured limit.
//
// The limit must not be modified while any goroutines in the group are active.
func (pb *parallelBatch) SetLimit(n int) {
	if n < 0 {
		pb.sem = nil
		return
	}
	if len(pb.sem) != 0 {
		panic(fmt.Errorf("errgroup: modify limit while %v goroutines in the group are still active", len(pb.sem)))
	}
	pb.sem = make(chan parallelBatchToken, n)
}
