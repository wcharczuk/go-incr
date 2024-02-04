package incr

import (
	"context"
	"fmt"
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
func New(opts ...GraphOption) *Graph {
	options := GraphOptions{
		MaxRecomputeHeapHeight: DefaultMaxRecomputeHeapHeight,
	}
	for _, opt := range opts {
		opt(&options)
	}
	g := &Graph{
		id:                       NewIdentifier(),
		stabilizationNum:         1,
		status:                   StatusNotStabilizing,
		observed:                 newNodeList(),
		observers:                make(map[Identifier]IObserver),
		recomputeHeap:            newRecomputeHeap(options.MaxRecomputeHeapHeight),
		adjustHeightsList:        newNodeList(),
		setDuringStabilization:   newNodeList(),
		handleAfterStabilization: make(map[Identifier][]func(context.Context)),
	}
	return g
}

// GraphOption mutates GraphOptions.
type GraphOption func(*GraphOptions)

// GraphMaxRecomputeHeapHeight sets the graph max recompute height.
func GraphMaxRecomputeHeapHeight(maxHeight int) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.MaxRecomputeHeapHeight = maxHeight
	}
}

// GraphOptions are options for graphs.
type GraphOptions struct {
	MaxRecomputeHeapHeight int
}

const (
	// DefaultMaxRecomputeHeapHeight is the default maximum height that can
	// be tracked in the recompute heap.
	DefaultMaxRecomputeHeapHeight = 256
)

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
	observed *nodeList
	// observers hold references to observers organized by node id.
	observers map[Identifier]IObserver
	// observersMu interlocks acces to observers
	observersMu sync.Mutex

	// recomputeHeap is the heap of nodes to be processed
	// organized by pseudo-height. The recompute heap
	// itself is concurrent safe.
	recomputeHeap *recomputeHeap

	// adjustHeightsList is a list of nodes that are
	// queued to have their heights recalculated after
	// a level of the recompute heap is recomputed
	adjustHeightsList *nodeList

	// setDuringStabilization is a list of nodes that were
	// set during stabilization
	setDuringStabilization *nodeList

	// handleAfterStabilization is a list of update
	// handlers that need to run after stabilization is done.
	handleAfterStabilization map[Identifier][]func(context.Context)
	// handleAfterStabilizationMu coordinates access to handleAfterStabilization
	handleAfterStabilizationMu sync.Mutex

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
	ok = graph.observed.Has(gn)
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
// Internal discovery & observe methods
//

// observeNodes traverses up from a given node, adding a given
// list of observers as "observing" that node, and recursing through it's inputs or parents.
func (graph *Graph) observeNodes(ctx context.Context, gn INode, observers ...IObserver) {
	graph.observeSingleNode(ctx, gn, observers...)
	gnn := gn.Node()
	parents := gnn.Parents()
	for _, p := range parents {
		graph.observeNodes(ctx, p, observers...)
	}
}

func (graph *Graph) observeSingleNode(ctx context.Context, gn INode, observers ...IObserver) {
	gnn := gn.Node()

	gnn.addObservers(observers...)
	alreadyObservedByGraph := graph.maybeAddObservedNode(gn)
	if alreadyObservedByGraph {
		return
	}

	TracePrintf(ctx, "observing node %v", gn)
	graph.numNodes++
	gnn.graph = graph
	gnn.detectCutoff(gn)
	gnn.detectAlways(gn)
	gnn.detectStabilize(gn)
	if gnn.ShouldRecompute() {
		graph.recomputeHeap.Add(gn)
	}
}

func (graph *Graph) maybeAddObservedNode(gn INode) bool {
	graph.observed.Lock()
	defer graph.observed.Unlock()
	if graph.observed.HasUnsafe(gn) {
		return true
	}
	graph.observed.PushUnsafe(gn)
	return false
}

func (graph *Graph) unobserveNodes(ctx context.Context, gn INode, observers ...IObserver) {
	graph.unobserveSingleNode(ctx, gn, observers...)
	gnn := gn.Node()
	parents := gnn.Parents()
	for _, p := range parents {
		graph.unobserveNodes(ctx, p, observers...)
	}
}

func (graph *Graph) unobserveSingleNode(ctx context.Context, gn INode, observers ...IObserver) {
	gnn := gn.Node()

	// nodes may implement custom observe/unobserve
	// steps so we check for that here and call
	// their handler if they implement the interface.
	if typed, ok := gn.(IUnobserve); ok {
		typed.Unobserve(ctx)
	}

	remainingObserverCount := graph.removeNodeObservers(ctx, gn, observers...)
	if remainingObserverCount > 0 {
		return
	}
	gnn.graph = nil
	graph.observed.Remove(gn)
	graph.numNodes--
	graph.recomputeHeap.Remove(gn)
	graph.adjustHeightsList.Remove(gn)

	graph.handleAfterStabilizationMu.Lock()
	delete(graph.handleAfterStabilization, gn.Node().ID())
	graph.handleAfterStabilizationMu.Unlock()
}

func (graph *Graph) removeNodeObservers(ctx context.Context, gn INode, observers ...IObserver) (remainingObserverCount int) {
	gnn := gn.Node()
	gnn.observersMu.Lock()
	defer gnn.observersMu.Unlock()
	for _, on := range observers {
		delete(gnn.observers, on.Node().id)
		for _, handler := range gnn.onUnobservedHandlers {
			handler(on)
		}
	}
	remainingObserverCount = len(gnn.observers)
	return
}

func (graph *Graph) addObserver(ctx context.Context, on IObserver) {
	graph.observersMu.Lock()
	defer graph.observersMu.Unlock()

	onn := on.Node()
	onn.graph = graph
	if _, ok := graph.observers[onn.id]; !ok {
		graph.numNodes++
	}
	onn.detectStabilize(on)
	graph.observers[onn.id] = on
}

func (graph *Graph) removeObserver(ctx context.Context, on IObserver) {
	onn := on.Node()
	onn.graph = nil
	graph.numNodes--
	graph.recomputeHeap.Remove(on)

	graph.handleAfterStabilizationMu.Lock()
	delete(graph.handleAfterStabilization, on.Node().ID())
	graph.handleAfterStabilizationMu.Unlock()

	graph.observersMu.Lock()
	delete(graph.observers, onn.id)
	graph.observersMu.Unlock()
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
	graph.setDuringStabilization.Lock()
	defer graph.setDuringStabilization.Unlock()
	for !graph.setDuringStabilization.IsEmptyUnsafe() {
		nodes := graph.setDuringStabilization.PopAllUnsafe()
		for _, n := range nodes {
			_ = n.Node().maybeStabilize(ctx)
			graph.SetStale(n)
		}
	}
}

func (graph *Graph) stabilizeEndRunUpdateHandlers(ctx context.Context) {
	graph.handleAfterStabilizationMu.Lock()
	defer graph.handleAfterStabilizationMu.Unlock()

	atomic.StoreInt32(&graph.status, StatusRunningUpdateHandlers)
	if len(graph.handleAfterStabilization) > 0 {
		TracePrintln(ctx, "stabilization calling user update handlers starting")
		defer func() {
			TracePrintln(ctx, "stabilization calling user update handlers complete")
		}()
	}
	for _, uhGroup := range graph.handleAfterStabilization {
		for _, uh := range uhGroup {
			uh(ctx)
		}
	}
	clear(graph.handleAfterStabilization)
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
		TracePrintf(ctx, "stabilization saw cutoff node %v", n)
		return
	}

	TracePrintf(ctx, "stabilization is recomputing %v", n)
	graph.numNodesChanged++
	nn.numChanges++

	// we have to propagate the "changed" or "recomputed" status to children
	nn.changedAt = graph.stabilizationNum
	if err = nn.maybeStabilize(ctx); err != nil {
		for _, eh := range nn.onErrorHandlers {
			eh(ctx, err)
		}
		return
	}

	if len(nn.onUpdateHandlers) > 0 {
		graph.handleAfterStabilization[nn.id] = append(graph.handleAfterStabilization[nn.id], nn.onUpdateHandlers...)
	}

	// iterate over each child node of this node, or nodes that take
	// this node as an input, and if the node is "necessary" and
	// dirty add it to the recompute heap.
	//
	// we use the `each` from here to hold the lock while we process
	// the list, preventing a race condition around missed nodes,
	// but more to prevent us from reading the children list twice.
	nn.children.Each(func(c INode) {
		isObserving := graph.IsObserving(c)
		shouldRecompute := c.Node().ShouldRecompute()
		if isObserving && shouldRecompute {
			graph.recomputeHeap.Add(c)
		}
	})
	return
}

//
// internal height management methods
//

func (graph *Graph) fixAdjustHeightsList() {
	if graph.adjustHeightsList.Len() > 0 {
		graph.recomputeHeap.mu.Lock()
		defer graph.recomputeHeap.mu.Unlock()
		graph.adjustHeightsList.ConsumeEach(func(n INode) {
			graph.recomputeHeap.fixUnsafe(n.Node().id)
		})
	}
}
