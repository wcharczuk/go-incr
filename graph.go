package incr

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// New returns a new [Graph], which is the type that represents the
// shared state of a computation graph.
//
// You can pass configuration options as [GraphOption] to customize settings
// within the graph, such as what the maximum "height" a node can be.
//
// This is the entrypoint for all stabilization and computation
// operations, and generally the [Graph] will be passed to node constructors.
//
// Nodes you initialize the graph with will need to be be observed by
// an [Observer] before you can stabilize them.
func New(opts ...GraphOption) *Graph {
	options := GraphOptions{
		MaxHeight:   DefaultMaxHeight,
		Parallelism: runtime.NumCPU(),
	}
	for _, opt := range opts {
		opt(&options)
	}
	return &Graph{
		id:                        NewIdentifier(),
		parallelism:               options.Parallelism,
		clearRecomputeHeapOnError: options.ClearRecomputeHeapOnError,
		stabilizationNum:          1,
		status:                    StatusNotStabilizing,
		nodes:                     allocateMapWithSize[Identifier, INode](options.PreallocateNodesSize),
		observers:                 allocateMapWithSize[Identifier, IObserver](options.PreallocateObserversSize),
		sentinels:                 allocateMapWithSize[Identifier, ISentinel](options.PreallocateSentinelsSize),
		recomputeHeap:             newRecomputeHeap(options.MaxHeight),
		adjustHeightsHeap:         newAdjustHeightsHeap(options.MaxHeight),
		setDuringStabilization:    make(map[Identifier]INode),
		handleAfterStabilization:  make(map[Identifier][]func(context.Context)),
		propagateInvalidityQueue:  new(queue[INode]),
	}
}

func allocateMapWithSize[K comparable, V any](size int) map[K]V {
	if size > 0 {
		return make(map[K]V, size)
	}
	return make(map[K]V)
}

// GraphOption mutates GraphOptions.
type GraphOption func(*GraphOptions)

// OptGraphMaxHeight sets the graph max recompute height.
func OptGraphMaxHeight(maxHeight int) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.MaxHeight = maxHeight
	}
}

// OptGraphParallelism sets the parallelism factor, or said another way
// the number of goroutines, to use when stabilizing using [Graph.ParallelStabilize].
//
// This will default to [runtime.NumCPU] if unset.
func OptGraphParallelism(parallelism int) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.Parallelism = parallelism
	}
}

// OptGraphPreallocateNodesSize preallocates the node tracking map within
// the graph with a given size number of elements for items.
//
// If not provided, no size for elements will be preallocated.
func OptGraphPreallocateNodesSize(size int) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.PreallocateNodesSize = size
	}
}

// OptGraphPreallocateObserversSize preallocates the observer tracking map within
// the graph with a given size number of elements for items.
//
// If not provided, no size for elements will be preallocated.
func OptGraphPreallocateObserversSize(size int) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.PreallocateObserversSize = size
	}
}

// OptGraphPreallocateSentinelsSize preallocates the sentinel tracking map within
// the graph with a given size number of elements for items.
//
// If not provided, no size for elements will be preallocated.
func OptGraphPreallocateSentinelsSize(size int) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.PreallocateSentinelsSize = size
	}
}

// OptGraphClearRecomputeHeapOnError controls a setting for whether or not the
// recompute heap is cleared of nodes on stabilization error.
//
// If not provided, the default is to not clear the recompute heap, but leave nodes in place.
func OptGraphClearRecomputeHeapOnError(shouldClear bool) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.ClearRecomputeHeapOnError = shouldClear
	}
}

// GraphOptions are options for graphs.
type GraphOptions struct {
	MaxHeight                 int
	Parallelism               int
	PreallocateNodesSize      int
	PreallocateObserversSize  int
	PreallocateSentinelsSize  int
	ClearRecomputeHeapOnError bool
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
// You should instantiate this type with the [New] function.
//
// The graph holds information such as, how many stabilizations have happened,
// what node are currently observed, and what nodes need to be recomputed.
type Graph struct {
	// id is a unique identifier for the graph
	id Identifier
	// label is a descriptive label for the graph
	label string

	// parallelism is the degree of parallelism used when processing nodes
	// with the [parallelBatch] iterator.
	parallelism int

	// clearRecomputeHeapOnError controls if we should clear the recomputeHeap on error.
	clearRecomputeHeapOnError bool

	// nodesMu interlocks access to nodes
	nodesMu sync.Mutex
	// observed are the nodes that the graph currently observes
	// organized by node id.
	nodes map[Identifier]INode

	// observersMu interlocks access to observers
	observersMu sync.Mutex
	// observers hold references to observers organized by node id.
	observers map[Identifier]IObserver

	// sentinelsMu interlocks access to sentinels
	sentinelsMu sync.Mutex
	// sentinels hold references to sentinels organized by node id.
	sentinels map[Identifier]ISentinel

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

	propagateInvalidityQueue *queue[INode]
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
func (graph *Graph) HasObserver(on IObserver) (ok bool) {
	graph.observersMu.Lock()
	_, ok = graph.observers[on.Node().id]
	graph.observersMu.Unlock()
	return
}

// HasSentinel returns if a graph has a given sentinel.
func (graph *Graph) HasSentinel(sn ISentinel) (ok bool) {
	graph.sentinelsMu.Lock()
	_, ok = graph.sentinels[sn.Node().id]
	graph.sentinelsMu.Unlock()
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
	if gn.Node().heightInRecomputeHeap == HeightUnset {
		graph.recomputeHeap.add(gn)
	}
}

//
// Scope interface methods
//

func (graph *Graph) isTopScope() bool       { return true }
func (graph *Graph) isScopeValid() bool     { return true }
func (graph *Graph) isScopeNecessary() bool { return true }
func (graph *Graph) scopeGraph() *Graph     { return graph }
func (graph *Graph) scopeHeight() int       { return HeightUnset }
func (graph *Graph) addScopeNode(_ INode)   {}
func (graph *Graph) String() string         { return fmt.Sprintf("{graph:%s}", graph.id.Short()) }

//
// Internal discovery & observe methods
//

func (graph *Graph) invalidateNode(node INode) {
	if !node.Node().valid {
		return
	}

	nn := node.Node()
	nn.changedAt = graph.stabilizationNum
	nn.recomputedAt = graph.stabilizationNum
	if nn.isNecessary() {
		graph.removeParents(node)
		nn.height = node.Node().createdIn.scopeHeight() + 1
	}
	nn.maybeInvalidate()
	nn.valid = false
	for _, child := range node.Node().children {
		graph.propagateInvalidityQueue.push(child)
	}
	if node.Node().heightInRecomputeHeap != HeightUnset {
		graph.recomputeHeap.remove(node)
	}
}

func (graph *Graph) removeParents(child INode) {
	for _, parent := range child.Node().nodeParents() {
		graph.removeParent(child, parent)
	}
}

func (graph *Graph) unlink(child, parent INode) {
	child.Node().removeParent(parent.Node().id)
	parent.Node().removeChild(child.Node().id)
}

func (graph *Graph) removeParent(child, parent INode) {
	graph.unlink(child, parent)
	graph.checkIfUnnecessary(parent)
}

func (graph *Graph) checkIfUnnecessary(parent INode) {
	if !parent.Node().isNecessary() {
		graph.becameUnnecessary(parent)
	}
}

func (graph *Graph) becameUnnecessary(parent INode) {
	graph.removeParents(parent)
	graph.removeNode(parent)
}

func (graph *Graph) edgeIsStale(child, parent INode) bool {
	return parent.Node().changedAt > child.Node().recomputedAt
}

var errChildNil = errors.New("child node is <nil>, cannot continue")
var errParentNil = errors.New("parent node is <nil>, cannot continue")

func (graph *Graph) addChild(child, parent INode) error {
	if child == nil {
		return errChildNil
	}
	if parent == nil {
		return errParentNil
	}
	if err := graph.addChildWithoutAdjustingHeights(child, parent); err != nil {
		return err
	}
	if parent.Node().height >= child.Node().height {
		if err := graph.adjustHeightsHeap.adjustHeights(graph.recomputeHeap, child, parent); err != nil {
			return err
		}
	}
	graph.propagateInvalidity()
	if child.Node().recomputedAt == 0 || graph.edgeIsStale(child, parent) {
		graph.recomputeHeap.addIfNotPresent(child)
	}
	return nil
}

func (graph *Graph) changeParent(child, oldParent, newParent INode) error {
	if oldParent != nil && newParent != nil {
		if oldParent.Node().id == newParent.Node().id {
			return nil
		}
		oldParent.Node().removeChild(child.Node().id)
		oldParent.Node().forceNecessary = true
		if err := graph.addChild(child, newParent); err != nil {
			return err
		}
		oldParent.Node().forceNecessary = false
		graph.checkIfUnnecessary(oldParent)
		return nil
	}
	if oldParent == nil {
		return graph.addChild(child, newParent)
	}

	// newParent is nil
	oldParent.Node().removeChild(child.Node().id)
	graph.checkIfUnnecessary(oldParent)
	return nil
}

func (graph *Graph) propagateInvalidity() {
	for graph.propagateInvalidityQueue.len() > 0 {
		node, _ := graph.propagateInvalidityQueue.pop()
		if node.Node().valid {
			if node.Node().shouldBeInvalidated() {
				graph.invalidateNode(node)
			} else {
				graph.recomputeHeap.addIfNotPresent(node)
			}
		}
	}
}

func (graph *Graph) link(child, parent INode) {
	parent.Node().addChildren(child)
	child.Node().addParents(parent)
}

func (graph *Graph) addChildWithoutAdjustingHeights(child, parent INode) error {
	wasNecessary := parent.Node().isNecessary()
	graph.link(child, parent)
	if !parent.Node().valid {
		graph.propagateInvalidityQueue.push(child)
	}
	if !wasNecessary {
		if err := graph.becameNecessaryRecursive(parent); err != nil {
			return err
		}
	}
	return nil
}

func (graph *Graph) becameNecessaryRecursive(node INode) (err error) {
	graph.addNode(node)
	if err = graph.adjustHeightsHeap.setHeight(node, node.Node().createdIn.scopeHeight()+1); err != nil {
		return
	}
	for _, parent := range node.Node().nodeParents() {
		if err = graph.addChildWithoutAdjustingHeights(node, parent); err != nil {
			return err
		}
		if parent.Node().height >= node.Node().height {
			if err = graph.adjustHeightsHeap.setHeight(node, parent.Node().height+1); err != nil {
				return
			}
		}
	}
	for _, sentinels := range node.Node().sentinels {
		graph.recomputeHeap.addIfNotPresent(sentinels)
	}
	if node.Node().isStale() {
		graph.recomputeHeap.addIfNotPresent(node)
	}
	return
}

func (graph *Graph) becameNecessary(node INode) error {
	if err := graph.becameNecessaryRecursive(node); err != nil {
		return err
	}
	graph.propagateInvalidity()
	return nil
}

func (graph *Graph) addNode(n INode) {
	graph.nodesMu.Lock()
	defer graph.nodesMu.Unlock()

	gnn := n.Node()
	_, graphAlreadyHasNode := graph.nodes[gnn.id]
	if graphAlreadyHasNode {
		return
	}
	graph.numNodes++
	gnn.initializeFrom(n)
	graph.nodes[gnn.id] = n
}

func (graph *Graph) addObserver(on IObserver) {
	graph.observersMu.Lock()
	defer graph.observersMu.Unlock()

	onn := on.Node()
	_, graphAlreadyHasObserver := graph.observers[onn.id]
	if graphAlreadyHasObserver {
		return
	}
	graph.numNodes++
	onn.initializeFrom(on)
	graph.observers[onn.id] = on
}

func (graph *Graph) addSentinel(sn ISentinel) {
	graph.sentinelsMu.Lock()
	defer graph.sentinelsMu.Unlock()

	snn := sn.Node()
	_, graphAlreadyHasSentinel := graph.sentinels[snn.id]
	if graphAlreadyHasSentinel {
		return
	}
	graph.numNodes++
	snn.initializeFrom(sn)
	graph.sentinels[snn.id] = sn
}

func (graph *Graph) removeObserver(on IObserver) {
	graph.observersMu.Lock()
	delete(graph.observers, on.Node().id)
	graph.observersMu.Unlock()
	graph.zeroNode(on)
}

func (graph *Graph) removeSentinel(sn ISentinel) {
	graph.sentinelsMu.Lock()
	delete(graph.sentinels, sn.Node().id)
	graph.sentinelsMu.Unlock()
	graph.zeroNode(sn)
}

func (graph *Graph) removeNode(gn INode) {
	graph.nodesMu.Lock()
	delete(graph.nodes, gn.Node().id)
	graph.nodesMu.Unlock()
	graph.zeroNode(gn)
}

func (graph *Graph) zeroNode(n INode) {
	if n.Node().heightInRecomputeHeap != HeightUnset {
		graph.recomputeHeap.remove(n)
	}

	graph.numNodes--

	nn := n.Node()

	graph.handleAfterStabilizationMu.Lock()
	delete(graph.handleAfterStabilization, nn.ID())
	graph.handleAfterStabilizationMu.Unlock()

	nn.setAt = 0
	nn.changedAt = 0
	nn.recomputedAt = 0

	// mirror how we initialized the node
	nn.valid = true

	nn.parents = nil
	nn.children = nil
	nn.observers = nil

	// TODO (wc): why can't i zero these out?
	// nn.createdIn = nil
	nn.height = HeightUnset
	nn.heightInRecomputeHeap = HeightUnset
	nn.heightInAdjustHeightsHeap = HeightUnset
}

func (graph *Graph) observeNode(o IObserver, input INode) error {
	graph.addObserver(o)
	wasNecsesary := input.Node().isNecessary()
	input.Node().addObservers(o)
	if !wasNecsesary {
		if err := graph.becameNecessary(input); err != nil {
			return err
		}
	}
	graph.handleAfterStabilizationMu.Lock()
	graph.handleAfterStabilization[o.Node().id] = o.Node().onUpdateHandlers
	graph.handleAfterStabilizationMu.Unlock()
	return nil
}

func (graph *Graph) watchNode(sn ISentinel, input INode) error {
	graph.addSentinel(sn)
	input.Node().addSentinels(sn)
	graph.link(input, sn)
	if err := graph.adjustHeightsHeap.setHeight(sn, sn.Node().createdIn.scopeHeight()+1); err != nil {
		return err
	}
	return nil
}

func (graph *Graph) unobserveNode(o IObserver, input INode) {
	graph.removeObserver(o)
	input.Node().removeObserver(o.Node().id)
	graph.checkIfUnnecessary(input)
}

func (graph *Graph) unwatchNode(sn ISentinel, input INode) {
	graph.removeSentinel(sn)
	input.Node().removeSentinel(sn.Node().id)
	graph.unlink(input, sn)
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
func (graph *Graph) recompute(ctx context.Context, n INode, parallel bool) (err error) {
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
		return
	}

	graph.numNodesChanged++
	nn.numChanges++

	if err = nn.maybeStabilize(ctx); err != nil {
		for _, eh := range nn.onErrorHandlers {
			eh(ctx, err)
		}
		return
	}

	nn.changedAt = graph.stabilizationNum
	if len(nn.onUpdateHandlers) > 0 {
		graph.handleAfterStabilizationMu.Lock()
		graph.handleAfterStabilization[nn.id] = nn.onUpdateHandlers
		graph.handleAfterStabilizationMu.Unlock()
	}

	if parallel {
		graph.recomputeHeap.mu.Lock()
		for _, c := range nn.children {
			isNecessary := c.Node().isNecessary()
			isStale := c.Node().isStale()
			isNotInRecomputeHeap := c.Node().heightInRecomputeHeap == HeightUnset
			if isNecessary && isStale && isNotInRecomputeHeap {
				graph.recomputeHeap.addNodeUnsafe(c)
			}
		}
		graph.recomputeHeap.mu.Unlock()
	} else {
		for _, c := range nn.children {
			isNecessary := c.Node().isNecessary()
			isStale := c.Node().isStale()
			isNotInRecomputeHeap := c.Node().heightInRecomputeHeap == HeightUnset
			if isNecessary && isStale && isNotInRecomputeHeap {
				graph.recomputeHeap.addNodeUnsafe(c)
			}
		}
	}

	// recompute observers immediately because logically they're
	// children of this node but will not have any children themselves.
	for _, o := range nn.observers {
		if len(o.Node().onUpdateHandlers) > 0 {
			graph.handleAfterStabilizationMu.Lock()
			graph.handleAfterStabilization[nn.id] = o.Node().onUpdateHandlers
			graph.handleAfterStabilizationMu.Unlock()
		}
	}
	return
}
