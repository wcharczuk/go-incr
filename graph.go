package incr

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// New returns a new graph state, which is the type that
// represents the shared state of a computation graph.
//
// This is the entrypoint for all stabilization and computation
// operations.
//
// Nodes you initialize the graph with will be "observed" before
// the graph is returned, saving that step later.
func New(observed ...INode) *Graph {
	g := &Graph{
		id:                       NewIdentifier(),
		stabilizationNum:         1,
		status:                   StatusNotStabilizing,
		observed:                 make(map[Identifier]INode),
		setDuringStabilization:   new(list[Identifier, INode]),
		handleAfterStabilization: new(list[Identifier, []func(context.Context)]),
		recomputeHeap:            newRecomputeHeap(defaultRecomputeHeapMaxHeight),
	}
	g.Observe(observed...)
	return g
}

// Graph is the state that is shared across nodes.
//
// You should instantiate this type with `New()`.
type Graph struct {
	// id is a unique identifier for the graph
	id Identifier
	// mu is a synchronizing mutex for the graph
	mu sync.Mutex
	// observed are the nodes that the graph currently observes
	// organized by node id.
	observed map[Identifier]INode
	// recomputeHeap is the heap of nodes to be processed
	// organized by pseudo-height
	recomputeHeap *recomputeHeap
	// setDuringStabilization is a list of nodes that were
	// set during stabilization
	setDuringStabilization *list[Identifier, INode]
	// handleAfterStabilization is a list of update
	// handlers that need to run after stabilization is done.
	handleAfterStabilization *list[Identifier, []func(context.Context)]
	// stabilizationNum is the version
	// of the graph in respect to when
	// nodes are considered stale or changed
	stabilizationNum uint64
	// status is the general status of the graph where
	// the possible states are:
	// - StatusNotStabilizing (default)
	// - StatusStabilizing
	// - StatusRunningUpdateHandlers
	status int32
	// stabilizationStarted is the time of the stabilization pass currently in progress
	stabilizationStarted time.Time
	// numNodes are the total number of nodes found during
	// discovery and is typically used for testing
	numNodes uint64
	// numNodesRecomputed is the total number of nodes
	// that have been recomputed in the graph's history
	// and is typically used in testing
	numNodesRecomputed uint64
	// numNodesChanged is the total number of nodes
	// that have been changed in the graph's history
	// and is typically used in testing
	numNodesChanged uint64
}

// Observe observes a given list of nodes.
//
// The observation process involves discovering nodes
// linked to by dependency relationships of the given nodes
// setting up the computation metadata and other infrastructure
// to enable us to stabilize the computation later.
//
// Each node should in effect represent separate "graphs" but
// deduplication will be handled during the adding process.
func (graph *Graph) Observe(nodes ...INode) {
	graph.mu.Lock()
	for _, n := range nodes {
		graph.discoverAllNodes(n)
	}
	graph.mu.Unlock()
}

// IsObserving returns if a graph is observing a given node.
func (graph *Graph) IsObserving(gn INode) (ok bool) {
	_, ok = graph.observed[gn.Node().id]
	return
}

// SetStale sets a node as stale.
func (graph *Graph) SetStale(gn INode) {
	n := gn.Node()
	n.setAt = graph.stabilizationNum
	graph.recomputeHeap.Add(gn)
}

//
// internal methods
//

func (graph *Graph) discoverAllNodes(gn INode) {
	graph.discoverNode(gn)
	gnn := gn.Node()
	for _, c := range gnn.children {
		if graph.IsObserving(c) {
			continue
		}
		graph.discoverAllNodes(c)
	}
	for _, p := range gnn.parents {
		if graph.IsObserving(p) {
			continue
		}
		graph.discoverAllNodes(p)
	}
}

func (graph *Graph) discoverNode(gn INode) {
	graph.observed[gn.Node().id] = gn
	gnn := gn.Node()
	gnn.graph = graph
	gnn.detectCutoff(gn)
	gnn.detectStabilize(gn)
	gnn.height = gnn.pseudoHeight()
	graph.numNodes++
	graph.recomputeHeap.Add(gn)
	return
}

// undiscoverAllNodes removes a node and all its parents
// from a given graph.
//
// NOTE: you _must_ unlink it first or you'll just blow away the whole graph.
func (graph *Graph) undiscoverAllNodes(gn INode) {
	graph.undiscoverNode(gn)
	gnn := gn.Node()
	for _, c := range gnn.children {
		if !graph.IsObserving(c) {
			continue
		}
		graph.undiscoverAllNodes(c)
	}
	for _, p := range gnn.parents {
		if !graph.IsObserving(p) {
			continue
		}
		graph.undiscoverAllNodes(p)
	}
}

func (graph *Graph) undiscoverNode(gn INode) {
	gnn := gn.Node()
	gnn.graph = nil
	delete(graph.observed, gnn.id)
	graph.numNodes--
	graph.recomputeHeap.Remove(gn)
	graph.handleAfterStabilization.Remove(gn.Node().ID())
}

//
// stabilization methods
//

func (graph *Graph) ensureNotStabilizing(ctx context.Context) error {
	if atomic.LoadInt32(&graph.status) != StatusNotStabilizing {
		tracePrintf(ctx, "stabilize; already stabilizing, cannot continue")
		return ErrAlreadyStabilizing
	}
	return nil
}

func (graph *Graph) stabilizeStart(ctx context.Context) {
	atomic.StoreInt32(&graph.status, StatusStabilizing)
	graph.stabilizationStarted = time.Now()
	tracePrintf(ctx, "stabilize[%d]; stabilization starting", graph.stabilizationNum)
}

func (graph *Graph) stabilizeEnd(ctx context.Context) {
	defer func() {
		graph.stabilizationStarted = time.Time{}
		atomic.StoreInt32(&graph.status, StatusNotStabilizing)
	}()
	tracePrintf(ctx, "stabilize[%d]; stabilization complete (%v)", graph.stabilizationNum, time.Since(graph.stabilizationStarted).Round(time.Microsecond))
	graph.stabilizationNum++
	var n INode
	for !graph.setDuringStabilization.IsEmpty() {
		_, n, _ = graph.setDuringStabilization.Pop()
		_ = n.Node().maybeStabilize(ctx)
		graph.SetStale(n)
	}
	atomic.StoreInt32(&graph.status, StatusRunningUpdateHandlers)
	var updateHandlers []func(context.Context)
	if !graph.handleAfterStabilization.IsEmpty() {
		tracePrintf(ctx, "stabilize[%d]; calling update handlers starting", graph.stabilizationNum)
		defer func() {
			tracePrintf(ctx, "stabilize[%d]; calling update handlers complete", graph.stabilizationNum)
		}()
	}
	for !graph.handleAfterStabilization.IsEmpty() {
		_, updateHandlers, _ = graph.handleAfterStabilization.Pop()
		for _, uh := range updateHandlers {
			uh(ctx)
		}
	}
	return
}

// recompute starts the recompute cycle for the node
// setting the recomputedAt field and possibly changing the value.
func (graph *Graph) recompute(ctx context.Context, n *Node) error {
	graph.numNodesRecomputed++
	n.numRecomputes++
	n.recomputedAt = graph.stabilizationNum
	return graph.maybeChangeValue(ctx, n)
}

// maybeChangeValue checks the cutoff, and calls the stabilization
// delegate if one is set, adding the nodes parents to the recompute heap
// if there are changes.
func (graph *Graph) maybeChangeValue(ctx context.Context, n *Node) (err error) {
	if n.maybeCutoff(ctx) {
		return
	}
	graph.numNodesChanged++
	n.numChanges++
	n.changedAt = graph.stabilizationNum
	if err = n.maybeStabilize(ctx); err != nil {
		for _, eh := range n.onErrorHandlers {
			eh(ctx, err)
		}
		return
	}
	if len(n.onUpdateHandlers) > 0 {
		graph.handleAfterStabilization.Push(n.id, n.onUpdateHandlers)
	}
	for _, p := range n.parents {
		if p.Node().shouldRecompute() {
			graph.recomputeHeap.Add(p)
		}
	}
	return
}
