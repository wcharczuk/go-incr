package incr

import (
	"context"
	"fmt"
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
		observers:                make(map[Identifier]IObserver),
		setDuringStabilization:   new(list[Identifier, INode]),
		handleAfterStabilization: new(list[Identifier, []func(context.Context)]),
		recomputeHeap:            newRecomputeHeap(256),
	}
	return g
}

// Graph is the state that is shared across nodes.
//
// You should instantiate this type with `New()`.
//
// It is important to note that most operations on the graph are _not_ concurrent
// safe and you should use your own mutex to synchronize access to internal state.
type Graph struct {
	// id is a unique identifier for the graph
	id Identifier
	// label is a descriptive label for the graph
	label string
	// observed are the nodes that the graph currently observes
	// organized by node id.
	observed map[Identifier]INode
	// observers hold references to observers organized by node id.
	observers map[Identifier]IObserver

	// recomputeHeap is the heap of nodes to be processed
	// organized by pseudo-height. The recompute heap
	// itself is concurrent safe.
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

	// metadata is extra data you can add to the graph instance and
	// manage yourself.
	metadata any

	// onStabilizationStart are optional hooks called when stabilization starts.
	onStabilizationStart []func(context.Context)

	// onStabilizationEnd are optional hooks called when stabilization ends.
	onStabilizationEnd []func(context.Context, time.Time, error)
}

// ID is the identifier for the graph.
func (graph *Graph) ID() Identifier {
	return graph.id
}

// Label returns the graph label.
func (graph *Graph) Label() string {
	return graph.label
}

// SetLabel sets the graph label.
func (graph *Graph) SetLabel(label string) {
	graph.label = label
}

// Metadata is extra data held on the graph instance.
func (graph *Graph) Metadata() any {
	return graph.metadata
}

// SetMetadata sets the metadata for the graph instance.
func (graph *Graph) SetMetadata(metadata any) {
	graph.metadata = metadata
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

// OnStabilizationStart adds a stabilization start handler.
func (graph *Graph) OnStabilizationStart(handler func(context.Context)) {
	graph.onStabilizationStart = append(graph.onStabilizationStart, handler)
}

// OnStabilizationEnd adds a stabilization end handler.
func (graph *Graph) OnStabilizationEnd(handler func(context.Context, time.Time, error)) {
	graph.onStabilizationEnd = append(graph.onStabilizationEnd, handler)
}

// Node helpers

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
	graph.discoverNodesContext(context.Background(), on, gn)
}

func (graph *Graph) DiscoverNode(on IObserver, gn INode) {
	graph.discoverNodeContext(context.Background(), on, gn)
}

// DiscoverObserver initializes an observer node
// which is treated specially by the graph.
func (graph *Graph) DiscoverObserver(on IObserver) {
	graph.discoverObserverContext(context.Background(), on)
}

// UndiscoverAllNodes removes a node and all its parents
// from a observation within a graph for a given observer.
func (graph *Graph) UndiscoverNodes(on IObserver, gn INode) {
	graph.undiscoverNodesContext(context.Background(), on, gn)
}

// UndiscoverNode removes the node from the graph
// observation lookup as well as updating internal
// stats metadata for the graph for a given observer.
func (graph *Graph) UndiscoverNode(on IObserver, gn INode) {
	graph.undiscoverNodeContext(context.Background(), on, gn)
}

// UndiscoverObserver removes an observer node
// which is treated specially by the graph.
func (graph *Graph) UndiscoverObserver(on IObserver) {
	graph.undiscoverObserverContext(context.Background(), on)
}

// RecomputeHeight recomputes the height of a given node.
//
// In practice, you should almost never need this function as
// heights are computed during observation, but in more
// mutable graph contexts it's helpful to trigger
// this step separately.
func (graph *Graph) RecomputeHeight(n INode) {
	nn := n.Node()
	nn.recomputeHeights()
}

//
// stabilization methods
//

func (graph *Graph) ensureNotStabilizing(ctx context.Context) error {
	if atomic.LoadInt32(&graph.status) != StatusNotStabilizing {
		TracePrintf(ctx, "stabilize; already stabilizing, cannot continue")
		return ErrAlreadyStabilizing
	}
	return nil
}

func (graph *Graph) stabilizeStart(ctx context.Context) context.Context {
	atomic.StoreInt32(&graph.status, StatusStabilizing)
	for _, handler := range graph.onStabilizationStart {
		handler(ctx)
	}
	graph.stabilizationStarted = time.Now()
	ctx = WithStabilizationNumber(ctx, graph.stabilizationNum)
	TracePrintln(ctx, "stabilization starting")
	return ctx
}

func (graph *Graph) stabilizeEnd(ctx context.Context, err error) {
	defer func() {
		graph.stabilizationStarted = time.Time{}
		atomic.StoreInt32(&graph.status, StatusNotStabilizing)
	}()
	for _, handler := range graph.onStabilizationEnd {
		handler(ctx, graph.stabilizationStarted, err)
	}
	if err != nil {
		TraceErrorf(ctx, "stabilization error: %v", err)
		TracePrintf(ctx, "stabilization failed (%v elapsed)", time.Since(graph.stabilizationStarted).Round(time.Microsecond))
	} else {
		TracePrintf(ctx, "stabilization complete (%v elapsed)", time.Since(graph.stabilizationStarted).Round(time.Microsecond))
	}
	graph.stabilizeEndRunUpdateHandlers(ctx)
	graph.stabilizationNum++
	graph.stabilizeEndHandleSetDuringStabilization(ctx)
}

func (graph *Graph) stabilizeEndHandleSetDuringStabilization(ctx context.Context) {
	graph.setDuringStabilization.mu.Lock()
	defer graph.setDuringStabilization.mu.Unlock()
	for !graph.setDuringStabilization.isEmptyUnsafe() {
		nodes := graph.setDuringStabilization.popAllUnsafe()
		for _, n := range nodes {
			_ = n.Node().maybeStabilize(ctx)
			graph.SetStale(n)
		}
	}
}

func (graph *Graph) stabilizeEndRunUpdateHandlers(ctx context.Context) {
	graph.handleAfterStabilization.mu.Lock()
	defer graph.handleAfterStabilization.mu.Unlock()

	atomic.StoreInt32(&graph.status, StatusRunningUpdateHandlers)

	if !graph.handleAfterStabilization.isEmptyUnsafe() {
		TracePrintln(ctx, "stabilization calling user update handlers starting")
		defer func() {
			TracePrintln(ctx, "stabilization calling user update handlers complete")
		}()
	}
	var updateHandlers [][]func(context.Context)
	for !graph.handleAfterStabilization.isEmptyUnsafe() {
		updateHandlers = graph.handleAfterStabilization.popAllUnsafe()
		for _, uhGroup := range updateHandlers {
			for _, uh := range uhGroup {
				uh(ctx)
			}
		}
	}
}

// recompute starts the recompute cycle for the node
// setting the recomputedAt field and possibly changing the value.
func (graph *Graph) recompute(ctx context.Context, n INode) (err error) {
	graph.numNodesRecomputed++

	nn := n.Node()
	if nn == nil {
		return fmt.Errorf("attempting to recompute uninitialized node; cannot continue")
	}

	nn.numRecomputes++
	nn.recomputedAt = graph.stabilizationNum

	// if the computation is aborted here don't proceed.
	var shouldCutoff bool
	shouldCutoff, err = nn.maybeCutoff(ctx)
	if err != nil {
		for _, eh := range nn.onErrorHandlers {
			eh(ctx, err)
		}
		return
	}
	if shouldCutoff {
		TracePrintf(ctx, "stabilization saw cutoff node %v with height %d", n, n.Node().height)
		return
	}

	TracePrintf(ctx, "stabilization is recomputing %v with height %d", n, n.Node().height)
	graph.numNodesChanged++
	nn.numChanges++

	// we have to propagate the "changed" or "recomputed" status to children
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
		graph.handleAfterStabilization.Push(nn.id, nn.onUpdateHandlers, 0)
	}

	// recompute all the children of this node, i.e. the nodes that
	// depend on this node if they need to be recomputed.
	for _, c := range nn.children {
		if graph.IsObserving(c) && c.Node().ShouldRecompute() {
			graph.recomputeHeap.Add(c)
		} else {
			TracePrintf(ctx, "stabilization is skipping recompute %v child %v", n, c)
		}
	}
	return
}

//
// internal discovery methods
//

// DiscoverNodes initializes tracking of a given node for a given observer
// and walks the nodes parents doing the same for any nodes seen.
func (graph *Graph) discoverNodesContext(ctx context.Context, on IObserver, gn INode) {
	graph.discoverNodeContext(ctx, on, gn)
	gnn := gn.Node()
	for _, p := range gnn.parents {
		graph.discoverNodesContext(ctx, on, p)
	}
}

// DiscoverNode initializes a node and adds
// it to the observed lookup.
func (graph *Graph) discoverNodeContext(ctx context.Context, on IObserver, gn INode) {
	TracePrintf(ctx, "discovering node %v", gn)
	gnn := gn.Node()
	nodeID := gnn.id

	// make sure to associate the given observer with the node
	gnn.observers[on.Node().ID()] = on
	for _, handler := range gnn.onObservedHandlers {
		handler(on)
	}

	// if the node is not currently observed.
	if _, ok := graph.observed[nodeID]; !ok {
		graph.numNodes++
	}
	graph.observed[nodeID] = gn
	gnn.graph = graph
	gnn.detectCutoff(gn)
	gnn.detectAlways(gn)
	gnn.detectStabilize(gn)
	gnn.detectBind(gn)
	gnn.height = gnn.computePseudoHeight()

	shouldRecompute := gnn.ShouldRecompute()
	recomputeHeapHasNode := graph.recomputeHeap.Has(gn)

	// we should add to the heap if we should recompute the node _or_ if we need to
	// potentially adjust the height it's sitting in the heap.
	if shouldRecompute || recomputeHeapHasNode {
		graph.recomputeHeap.Add(gn)
	}
}

func (graph *Graph) discoverObserverContext(ctx context.Context, on IObserver) {
	onn := on.Node()
	onn.graph = graph
	if _, ok := graph.observers[onn.id]; !ok {
		graph.numNodes++
	}
	onn.detectStabilize(on)
	graph.observers[onn.id] = on
	onn.height = onn.computePseudoHeight()
}

func (graph *Graph) undiscoverNodesContext(ctx context.Context, on IObserver, gn INode) {
	graph.undiscoverNodeContext(ctx, on, gn)
	gnn := gn.Node()
	for _, p := range gnn.parents {
		graph.undiscoverNodesContext(ctx, on, p)
	}
}

func (graph *Graph) undiscoverNodeContext(ctx context.Context, on IObserver, gn INode) {
	gnn := gn.Node()
	delete(gnn.observers, on.Node().ID())
	for _, handler := range gnn.onUnobservedHandlers {
		handler(on)
	}
	if len(gnn.observers) == 0 {
		TracePrintf(ctx, "undiscovering node %v; removing from observed", gn)
		gnn.graph = nil
		delete(graph.observed, gnn.id)
		graph.numNodes--
		graph.recomputeHeap.Remove(gn)
		graph.handleAfterStabilization.Remove(gn.Node().ID())
	} else {
		TracePrintf(ctx, "undiscovering node %v; remains observed", gn)
	}
}

func (graph *Graph) undiscoverObserverContext(ctx context.Context, on IObserver) {
	onn := on.Node()
	onn.graph = nil
	graph.numNodes--
	graph.recomputeHeap.Remove(on)
	graph.handleAfterStabilization.Remove(on.Node().ID())
}
