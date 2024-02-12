package incr

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// New returns a new graph state, which is the type that represents the
// shared state of a computation graph.
//
// You can pass configuration options as `GraphOption` to customize settings
// within the graph, such as what the maximum "height" a node can be.
//
// This is the entrypoint for all stabilization and computation
// operations, and generally the Graph will be passed to node constructors.
//
// Nodes you initialize the graph with will need to be be observed by
// an `Observer` before you can stabilize them.
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
		nodes:                    make(map[Identifier]INode),
		observers:                make(map[Identifier]IObserver),
		recomputeHeap:            newRecomputeHeap(options.MaxHeight),
		adjustHeightsHeap:        newAdjustHeightsHeap(options.MaxHeight),
		setDuringStabilization:   make(map[Identifier]INode),
		handleAfterStabilization: make(map[Identifier][]func(context.Context)),
	}
	return g
}

// GraphOption mutates GraphOptions.
type GraphOption func(*GraphOptions)

// GraphMaxRecomputeHeapHeight sets the graph max recompute height.
func GraphMaxRecomputeHeapHeight(maxHeight int) func(*GraphOptions) {
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
	// be tracked in the recompute heap.
	DefaultMaxHeight = 256
)

var (
	_ Scope = (*Graph)(nil)
)

// Graph is the state that is shared across nodes in a computation graph.
//
// You should instantiate this type with the `New()` function.
//
// The graph holds information such as, how many stabilizations have happened,
// what node are currently observed, and what nodes need to be recomputed.
type Graph struct {
	// id is a unique identifier for the graph
	id Identifier
	// label is a descriptive label for the graph
	label string

	nodesMu sync.Mutex
	// observed are the nodes that the graph currently observes
	// organized by node id.
	nodes map[Identifier]INode

	// observersMu interlocks acces to observers
	observersMu sync.Mutex
	// observers hold references to observers organized by node id.
	observers map[Identifier]IObserver

	// recomputeHeap is the heap of nodes to be processed
	// organized by pseudo-height. The recompute heap
	// itself is concurrent safe.
	recomputeHeap *recomputeHeap
	// adjustHeightsHeap is a list of nodes to adjust the heights for.
	adjustHeightsHeap *adjustHeightsHeap

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
func (graph *Graph) Has(gn INode) (ok bool) {
	graph.nodesMu.Lock()
	_, ok = graph.nodes[gn.Node().id]
	graph.nodesMu.Unlock()
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

// Node helpers

// SetStale sets a node as stale.
func (graph *Graph) SetStale(gn INode) {
	n := gn.Node()
	n.setAt = graph.stabilizationNum
	graph.recomputeHeap.add(gn)
}

//
// Scope interface methods
//

func (graph *Graph) isRootScope() bool { return true }

func (graph *Graph) scopeGraph() *Graph { return graph }

func (graph *Graph) scopeHeight() int { return -1 }

func (graph *Graph) String() string { return fmt.Sprintf("{graph:%s}", graph.id.Short()) }

//
// Internal discovery & observe methods
//

func (graph *Graph) removeParents(child INode) {
	for _, parent := range child.Node().parents {
		graph.removeParent(child, parent)
	}
}

func (graph *Graph) removeParent(child, parent INode) {
	Unlink(child, parent)
	graph.checkIfUnnecessary(parent)
}

func (graph *Graph) checkIfUnnecessary(n INode) {
	if !graph.isNecessary(n) {
		graph.removeNodeFromGraph(n)
		graph.removeParents(n)
	}
}

func (graph *Graph) becameNecessary(node INode) {
	graph.initializeNode(node)
	_ = graph.adjustHeightsHeap.setHeight(node, heightFromScope(node)+1)
	for _, p := range node.Node().parents {
		if p.Node().height >= node.Node().height {
			_ = graph.adjustHeightsHeap.setHeight(node, p.Node().height+1)
		}
		graph.becameNecessary(p)
	}

	if node.Node().ShouldRecompute() {
		graph.recomputeHeap.add(node)
	}
}

func (graph *Graph) isNecessary(n INode) bool {
	nn := n.Node()
	if _, isObserver := n.(IObserver); isObserver {
		return true
	}
	return len(nn.children) > 0 || len(nn.observers) > 0
}

func (graph *Graph) initializeNode(gn INode) {
	gnn := gn.Node()
	gnn.graph = graph
	graphAlreadyHasNode := graph.maybeAddNodeToGraph(gn)
	if graphAlreadyHasNode {
		return
	}
	graph.numNodes++
	gnn.detectCutoff(gn)
	gnn.detectAlways(gn)
	gnn.detectStabilize(gn)
}

func (graph *Graph) maybeAddNodeToGraph(gn INode) (ok bool) {
	graph.nodesMu.Lock()
	defer graph.nodesMu.Unlock()
	if _, ok = graph.nodes[gn.Node().id]; ok {
		return
	}
	graph.nodes[gn.Node().id] = gn
	return
}

func (graph *Graph) removeNodeFromGraph(gn INode) {
	graph.recomputeHeap.remove(gn)
	graph.adjustHeightsHeap.remove(gn)

	graph.nodesMu.Lock()
	delete(graph.nodes, gn.Node().id)
	graph.nodesMu.Unlock()

	graph.numNodes--

	gnn := gn.Node()

	graph.handleAfterStabilizationMu.Lock()
	delete(graph.handleAfterStabilization, gnn.ID())
	graph.handleAfterStabilizationMu.Unlock()

	gnn.setAt = 0
	gnn.boundAt = 0
	gnn.recomputedAt = 0

	// NOTE (wc): we never _really_ can remove the createdIn reference because
	// we don't track construction of nodes carefully.
	// gnn.createdIn = nil
	gnn.graph = nil
	gnn.height = 0
	gnn.heightInRecomputeHeap = 0
	gnn.heightInAdjustHeightsHeap = 0
}

func (graph *Graph) addObserver(on IObserver) error {
	onn := on.Node()
	onn.graph = graph

	graph.observersMu.Lock()
	if _, ok := graph.observers[onn.id]; !ok {
		graph.numNodes++
		graph.observers[onn.id] = on
	}
	graph.observersMu.Unlock()
	onn.detectStabilize(on)

	if err := graph.adjustHeights(on); err != nil {
		return err
	}
	return nil
}

func (graph *Graph) removeObserver(on IObserver) {
	onn := on.Node()
	onn.graph = nil
	graph.numNodes--
	graph.recomputeHeap.remove(on)

	graph.handleAfterStabilizationMu.Lock()
	delete(graph.handleAfterStabilization, on.Node().ID())
	graph.handleAfterStabilizationMu.Unlock()

	graph.observersMu.Lock()
	delete(graph.observers, onn.id)
	graph.observersMu.Unlock()

	graph.recomputeHeap.remove(on)
	graph.adjustHeightsHeap.remove(on)

	onn.height = 0
	onn.heightInRecomputeHeap = 0
	onn.heightInAdjustHeightsHeap = 0
	onn.setAt = 0
	onn.changedAt = 0
	onn.recomputedAt = 0
}

func (graph *Graph) adjustHeights(node INode) error {
	_ = graph.adjustHeightsHeap.setHeight(node, heightFromScope(node)+1)
	for _, p := range node.Node().parents {
		if p.Node().height >= node.Node().height {
			_ = graph.adjustHeightsHeap.setHeight(node, p.Node().height+1)
		}
	}
	for _, parent := range node.Node().parents {
		if err := graph.adjustHeightsHeap.adjustHeights(graph.recomputeHeap, node, parent); err != nil {
			return err
		}
	}
	return nil
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

// recompute starts the recompute cycle for the node
// setting the recomputedAt field and possibly changing the value.
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

	if err = nn.maybeStabilize(ctx); err != nil {
		for _, eh := range nn.onErrorHandlers {
			eh(ctx, err)
		}
		return
	}

	// we have to propagate the "changed" status to children
	nn.changedAt = graph.stabilizationNum

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
	for _, c := range nn.children {
		shouldRecompute := c.Node().ShouldRecompute()
		if graph.isNecessary(c) && shouldRecompute {
			graph.recomputeHeap.add(c)
		}
	}
	return
}
