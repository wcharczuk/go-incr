package incr

import (
	"context"
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
func New() *Graph {
	g := &Graph{
		id:                       NewIdentifier(),
		stabilizationNum:         1,
		status:                   StatusNotStabilizing,
		observed:                 make(map[Identifier]INode),
		setDuringStabilization:   new(list[Identifier, INode]),
		handleAfterStabilization: new(list[Identifier, []func(context.Context)]),
		recomputeHeap:            newRecomputeHeap(defaultRecomputeHeapMaxHeight),
	}
	return g
}

// Graph is the state that is shared across nodes.
//
// You should instantiate this type with `New()`.
type Graph struct {
	// id is a unique identifier for the graph
	id Identifier
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

// IsStabilizing returns if the graph is currently stabilizing.
func (graph *Graph) IsStabilizing() bool {
	return atomic.LoadInt32(&graph.status) != StatusNotStabilizing
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
// Discovery methods
//

// DiscoverNodes initializes tracking of a given node for a given observer
// and walks the nodes parents doing the same for any nodes seen.
func (graph *Graph) DiscoverNodes(on IObserver, gn INode) {
	graph.DiscoverNode(on, gn)
	gnn := gn.Node()
	for _, p := range gnn.parents {
		graph.DiscoverNodes(on, p)
	}
}

// DiscoverNode initializes a node and adds
// it to the observed lookup.
func (graph *Graph) DiscoverNode(on IObserver, gn INode) {
	gnn := gn.Node()
	nodeID := gnn.id

	for _, handler := range gnn.onObservedHandlers {
		handler(on)
	}
	gnn.observers = append(gnn.observers, on)

	// if the node is not currently observed.
	if _, ok := graph.observed[nodeID]; !ok {
		graph.observed[nodeID] = gn
		gnn.graph = graph
		gnn.detectCutoff(gn)
		gnn.detectStabilize(gn)
		gnn.detectBind(gn)
		graph.numNodes++
		gnn.height = gnn.computePseudoHeight()

		if gnn.ShouldRecompute() {
			graph.recomputeHeap.Add(gn)
		}
	}
	return
}

// DiscoverObserver initializes an observer node
// which is treated specially by the graph.
func (graph *Graph) DiscoverObserver(on IObserver) {
	onn := on.Node()
	onn.graph = graph
	graph.numNodes++
	onn.height = onn.computePseudoHeight()
	graph.recomputeHeap.Add(on)
	return
}

// UndiscoverAllNodes removes a node and all its parents
// from a observation within a graph for a given observer.
func (graph *Graph) UndiscoverNodes(on IObserver, gn INode) {
	graph.UndiscoverNode(on, gn)
	gnn := gn.Node()
	for _, c := range gnn.parents {
		if !graph.IsObserving(c) {
			continue
		}
		graph.UndiscoverNodes(on, c)
	}
}

// UndiscoverNode removes the node from the graph
// observation lookup as well as updating internal
// stats metadata for the graph for a given observer.
func (graph *Graph) UndiscoverNode(on IObserver, gn INode) {
	gnn := gn.Node()
	gnn.observers = filter(gnn.observers, func(on0 IObserver) bool {
		return on0.Node().id != on.Node().id
	})
	for _, handler := range gnn.onUnobservedHandlers {
		handler(on)
	}
	if len(gnn.observers) == 0 {
		gnn.graph = nil
		delete(graph.observed, gnn.id)
		graph.numNodes--
		graph.recomputeHeap.Remove(gn)
		graph.handleAfterStabilization.Remove(gn.Node().ID())
	}
}

// UndiscoverObserver removes an observer node
// which is treated specially by the graph.
func (graph *Graph) UndiscoverObserver(on IObserver) {
	onn := on.Node()
	onn.graph = nil
	graph.numNodes--
	graph.recomputeHeap.Remove(on)
	graph.handleAfterStabilization.Remove(on.Node().ID())
	return
}

// RecomputeHeight recomputes the height of a given node.
//
// In practice, you should almost never need this function as
// heights are computed during observation, but in more
// mutable graph contexts it's helpful to trigger
// this step separately.
func (graph *Graph) RecomputeHeight(n INode) {
	nn := n.Node()
	oldHeight := nn.height
	nn.recomputeHeights()
	if oldHeight != nn.height {
		if graph.recomputeHeap.Has(n) {
			graph.recomputeHeap.Add(n)
		}
	}
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
func (graph *Graph) recompute(ctx context.Context, n INode) (err error) {
	graph.numNodesRecomputed++

	nn := n.Node()
	nn.numRecomputes++
	nn.recomputedAt = graph.stabilizationNum
	if nn.maybeCutoff(ctx) {
		return
	}

	tracePrintf(ctx, "stabilize[%d]; recompute %v with height %d", graph.stabilizationNum, n, n.Node().height)
	graph.numNodesChanged++
	nn.numChanges++

	// we have to propagate the "changed" or "recomputed" status to parents
	nn.changedAt = graph.stabilizationNum
	if err = nn.maybeBind(ctx); err != nil {
		for _, eh := range nn.onErrorHandlers {
			eh(ctx, err)
		}
		return
	}
	if err = nn.maybeStabilize(ctx); err != nil {
		for _, eh := range nn.onErrorHandlers {
			eh(ctx, err)
		}
		return
	}
	if len(nn.onUpdateHandlers) > 0 {
		graph.handleAfterStabilization.Push(nn.id, nn.onUpdateHandlers)
	}

	// recompute all the children of this node
	// i.e. the nodes that depend on this node.
	for _, c := range nn.children {
		if c.Node().ShouldRecompute() {
			graph.recomputeHeap.Add(c)
		}
	}
	return
}
