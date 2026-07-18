package incr

import (
	"context"
)

// cancelCheckStride is how many nodes are recomputed between checks for a canceled
// context. Checking every node would add a measurable share of a cheap node's recompute
// to the hot loop; at this stride the cost disappears and an abort still lands within a
// couple of microseconds.
const cancelCheckStride = 64

// contextCanceled reports a context's cancellation cause, or nil while it is live.
//
// done is passed separately because a context that can never be canceled has a nil
// channel, which no select ever chooses -- so the check costs a branch rather than the
// atomic load that ctx.Err would do.
func contextCanceled(ctx context.Context, done <-chan struct{}) error {
	select {
	case <-done:
		return context.Cause(ctx)
	default:
		return nil
	}
}

// Stabilize kicks off the stabilization for nodes that have been observed by the graph's scope.
//
// The general process of stabilization is to scan the recompute heap up from the minimum height
// block, processing each node in the block until there are no more nodes.
//
// Stabilizing a node can add more nodes to the recompute heap, creating more work as the stabilization
// progresses, until finally no more nodes are left to process.
//
// The [Graph.Stabilize] stabilization process is serial, that is each node is recomputed in sequence one
// after the other.
//
// This can be extremely fast in practice because it lets us makes some assumptions about what
// can change in during each node's stabilization, specifically we can assume that [Bind] nodes
// evaluate serially and we can adjust recompute heights accordingly, and as a result we don't need
// to worry about shared resource contention and can skip acquiring locks.
//
// If during the stabilization pass a node's stabilize function returns an error, the recomputation pass
// is stopped and the error is returned. The node that failed is returned to the recompute
// heap, so that a later pass tries it again; a transient failure does not strand its value.
//
// Canceling ctx stops the pass at the next node boundary and returns the context's cause.
// Nodes not yet recomputed stay in the recompute heap, so a canceled pass leaves the same
// state as one stopped by an error, and stabilizing again continues from where it left off.
// A context that cannot be canceled -- [context.Background] and friends -- costs nothing.
func (graph *Graph) Stabilize(ctx context.Context) (err error) {
	if err = graph.ensureNotStabilizing(ctx); err != nil {
		return
	}
	ctx = graph.stabilizeStart(ctx)
	defer func() {
		graph.stabilizeEnd(ctx, err)
	}()
	// One guard for the whole pass, rather than one per node. A panic in user code becomes
	// an error: letting it unwind would abandon the pass with no record of which node was
	// responsible. The node is read from the graph because a panic unwinds past the loop
	// below before anything here could observe which node it was in.
	defer func() {
		if r := recover(); r != nil {
			err = graph.recomputePanicked(ctx, graph.recomputingNode, r)
			graph.handleStabilizationError(ctx, err)
		}
	}()

	var immediateRecompute []INode
	var next INode

	// Checking for cancellation costs about 4.5ns against a node recompute of about 25ns,
	// so it is done once per stride rather than per node, and not at all for a context that
	// can never be canceled -- which is the common case, and where the nil check below is
	// a branch that predicts perfectly. The stride starts spent, so a context that is
	// already canceled does no work at all.
	done := ctx.Done()
	cancellable := done != nil
	sinceCancelCheck := cancelCheckStride

	// note: this accesses the number of items directly
	// which is ~unsafe but we should only be able to
	// modify this if ensureNotStabilizing passes
	for graph.recomputeHeap.numItems > 0 {
		if cancellable {
			if sinceCancelCheck++; sinceCancelCheck >= cancelCheckStride {
				sinceCancelCheck = 0
				if err = contextCanceled(ctx, done); err != nil {
					break
				}
			}
		}
		next, _ = graph.recomputeHeap.removeMinUnsafe()
		err = graph.recompute(ctx, next, false /*parallel*/)
		if next.Node().always {
			immediateRecompute = append(immediateRecompute, next)
		}
		if err != nil {
			break
		}
	}
	if err != nil {
		graph.handleStabilizationError(ctx, err)
	}
	if len(immediateRecompute) > 0 {
		for _, n := range immediateRecompute {
			graph.recomputeHeap.addIfNotPresent(n)
		}
	}
	return
}

// handleStabilizationError does what a stopped pass leaves behind, for both an error
// returned by a node and a panic recovered from one.
func (graph *Graph) handleStabilizationError(ctx context.Context, err error) {
	if !graph.clearRecomputeHeapOnError {
		return
	}
	aborted := graph.recomputeHeap.clear()
	for _, node := range aborted {
		for _, ah := range node.Node().abortedHandlers() {
			ah(ctx, err)
		}
	}
}
