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
func New(opts ...GraphOption) *Graph {
	options := GraphOptions{
		MaxHeight: DefaultMaxHeight,
	}
	for _, opt := range opts {
		opt(&options)
	}
	g := &Graph{
		id:                       NewIdentifier(),
		stabilizationNum:         1,
		status:                   StatusNotStabilizing,
		observed:                 make(map[Identifier]INode),
		observers:                make(map[Identifier]IObserver),
		recomputeHeap:            newRecomputeHeap(options.MaxHeight),
		adjustHeightsHeap:        newAdjustHeightsHeap(options.MaxHeight),
		newObservers:             new(queue[IObserver]),
		propagateInvalidityQueue: new(queue[INode]),
		setDuringStabilization:   make(map[Identifier]INode),
		handleAfterStabilization: make(map[Identifier][]func(context.Context)),
	}
	return g
}

// GraphOption mutates GraphOptions.
type GraphOption func(*GraphOptions)

// GraphMaxHeight sets the graph max recompute height.
func GraphMaxHeight(maxHeight int) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.MaxHeight = maxHeight
	}
}

// GraphOptions are options for graphs.
type GraphOptions struct {
	MaxHeight int
}

const (
	// DefaultMaxHeight is the default maximum height that can
	// be tracked in the recompute and adjust heights heaps.
	DefaultMaxHeight = 256
)

var (
	_ Scope = (*Graph)(nil)
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

	observedMu sync.Mutex
	// observed are the nodes that the graph currently observes
	// organized by node id.
	observed map[Identifier]INode

	// observersMu interlocks acces to observers
	observersMu sync.Mutex
	// observers hold references to observers organized by node id.
	observers map[Identifier]IObserver

	// recomputeHeap is the heap of nodes to be processed
	// organized by pseudo-height. The recompute heap
	// itself is concurrent safe.
	recomputeHeap *recomputeHeap

	// adjustHeightsHeap is a heap of nodes
	// used to organize height calculations
	adjustHeightsHeap *adjustHeightsHeap

	// propagateInvalidityQueue holds nodes that have invalid children
	// that should be considered for invalidation.
	propagateInvalidityQueue *queue[INode]

	// newObsevers is a queue of recently registered observers
	newObservers *queue[IObserver]

	// setDuringStabilizationMu interlocks acces to setDuringStabilization
	setDuringStabilizationMu sync.Mutex
	// setDuringStabilization is a list of nodes that were
	// set during stabilization
	setDuringStabilization map[Identifier]INode

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
	graph.observedMu.Lock()
	_, ok = graph.observed[gn.Node().id]
	graph.observedMu.Unlock()
	return
}

// HasObserver returns if a graph has a given observer.
func (graph *Graph) HasObserver(gn INode) (ok bool) {
	graph.observersMu.Lock()
	_, ok = graph.observers[gn.Node().id]
	graph.observersMu.Unlock()
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

// Scope methods
func (graph *Graph) IsNecessary() bool { return true }
func (graph *Graph) IsValid() bool     { return true }
func (graph *Graph) AddNode(_ INode)   {}
func (graph *Graph) Height() int       { return heightUnset }
func (graph *Graph) IsTop() bool       { return true }
func (graph *Graph) Nodes() []INode    { return nil }
func (graph *Graph) Graph() *Graph     { return graph }

// Node helpers

// SetStale sets a node as stale.
func (graph *Graph) SetStale(gn INode) {
	n := gn.Node()
	n.setAt = graph.stabilizationNum
	graph.recomputeHeap.add(gn)
}

//
// Internal discovery & observe methods
//

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
	graph.setDuringStabilizationMu.Lock()
	defer graph.setDuringStabilizationMu.Unlock()
	for _, n := range graph.setDuringStabilization {
		_ = n.Node().maybeStabilize(ctx)
		graph.SetStale(n)
	}
	clear(graph.setDuringStabilization)
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

func (graph *Graph) recompute(ctx context.Context, n INode) (err error) {
	graph.numNodesRecomputed++
	nn := n.Node()
	nn.numRecomputes++
	nn.recomputedAt = graph.stabilizationNum

	var shouldCutoff bool
	shouldCutoff, err = nn.maybeCutoff(ctx)
	if err != nil {
		for _, eh := range nn.onErrorHandlers {
			eh(ctx, err)
		}
		return
	}
	if shouldCutoff {
		TracePrintf(ctx, "stabilization saw active cutoff %v", n)
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
	for _, p := range nn.parents {
		if p.Node().shouldRecompute() {
			graph.recomputeHeap.Add(p)
		}
	}
	return
}

//
// direct incremental translations
//
// The goal with this section is to transcribe (basically directly)
// the original intent of the incremental implementation's approach.
//

func (graph *Graph) link(child, parent INode) error {
	parent.Node().addChildren(child)
	child.Node().addParents(child)
	return nil
}
