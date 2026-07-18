package incr

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"
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
	if options.IdentifierProvider == nil {
		if options.Deterministic {
			options.IdentifierProvider = NewSequentialIdentifierProvier(1)
		} else {
			options.IdentifierProvider = _defaultIdentifierProvider
		}
	}
	return &Graph{
		identiferProvider:         options.IdentifierProvider,
		id:                        options.IdentifierProvider.NewIdentifier(),
		parallelism:               options.Parallelism,
		clearRecomputeHeapOnError: options.ClearRecomputeHeapOnError,
		deterministic:             options.Deterministic,
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
// By default the graph will not clear the recompute heap, and instead leave nodes in place.
//
// If this option is provided, and `shouldClear` is `true`, then the recompute heap
// will be cleared on error, and the `OnAborted` handlers of nodes will be called.
func OptGraphClearRecomputeHeapOnError(shouldClear bool) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.ClearRecomputeHeapOnError = shouldClear
	}
}

// OptGraphDeterministic ensures that processes the operate based on order do so consistently.
//
// If not provided, by default some processes, like calling update handlers after stabilization, use
// map iterators for which the execution order is undefined, though in practice it is pseudo-random.
func OptGraphDeterministic(deterministic bool) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.Deterministic = deterministic
	}
}

// OptGraphIdentifierProvider sets the graph's identifier provider.
//
// By default the graph will use a crypto/rand based identifier provider that returns
// identifiers using [NewCryptoRandIdentifierProvider].
//
// You can customize the identifier provider used by the graph for performance or determinism
// reasons.
func OptGraphIdentifierProvider(identifierProvider IdentifierProvider) func(*GraphOptions) {
	return func(g *GraphOptions) {
		g.IdentifierProvider = identifierProvider
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
	Deterministic             bool
	IdentifierProvider        IdentifierProvider
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
	// identifierProvider is the provider for identifiers
	identiferProvider IdentifierProvider
	// id is a unique identifier for the graph
	id Identifier
	// label is a descriptive label for the graph
	label string

	// parallelism is the degree of parallelism used when processing nodes
	// with the [parallelBatch] iterator.
	parallelism int

	// clearRecomputeHeapOnError controls if we should clear the recomputeHeap on error.
	clearRecomputeHeapOnError bool

	// deterministic controls aspects of stabilization such that
	// if the user values determinism, things will happen in consistent order.
	deterministic bool

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

	// recomputeMu serializes graph-structure mutation during
	// [Graph.ParallelStabilize]. Within a height block the parallel workers
	// run each node's user Stabilize concurrently, but the structural work
	// (a bind swapping out its subgraph, and enqueuing recomputed nodes'
	// children) must be serialized because it mutates shared node/height/heap
	// state. It is only ever contended on the parallel path; serial
	// stabilization is single-goroutine and never acquires it.
	recomputeMu sync.Mutex

	// setDuringStabilizationMu interlocks acces to setDuringStabilization
	setDuringStabilizationMu sync.Mutex
	// setDuringStabilization is a list of nodes that were
	// set during stabilization
	setDuringStabilization map[Identifier]INode

	// handleAfterStabilizationMu coordinates access to handleAfterStabilization
	handleAfterStabilizationMu sync.Mutex
	// handleAfterStabilization is a list of update
	// handlers that need to run after stabilization is done.
	handleAfterStabilization map[Identifier][]func(context.Context)

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
	// discovery and is typically used for testing.
	//
	// mutate it with the atomic helpers; nodes can be added and removed
	// concurrently during parallel stabilization (e.g. by binds at the
	// same height).
	numNodes uint64
	// numNodesRecomputed is the total number of nodes
	// that have been recomputed in the graph's history
	// and is typically used in testing.
	//
	// mutate it with the atomic helpers; nodes are recomputed
	// concurrently during parallel stabilization.
	numNodesRecomputed uint64
	// numNodesChanged is the total number of nodes
	// that have been changed in the graph's history
	// and is typically used in testing.
	//
	// mutate it with the atomic helpers; nodes are recomputed
	// concurrently during parallel stabilization.
	numNodesChanged uint64
	// numNodesRecomputedDirectly is the total number of nodes that were
	// recomputed immediately after the node they depend on, without being routed
	// through the recompute heap. See [Graph.canRecomputeImmediately].
	//
	// this only happens during serial stabilization, so it needs no atomics.
	numNodesRecomputedDirectly uint64

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
func (graph *Graph) newIdentifier() Identifier {
	return graph.identiferProvider.NewIdentifier()
}

//
// Internal discovery & observe methods
//

func (graph *Graph) invalidateNode(node INode) {
	for _, handler := range node.Node().invalidatedHandlers() {
		handler()
	}
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
	parents := child.Node().nodeParents()
	// A node can appear in a child's input list more than once, because a child
	// may take it as several of its inputs -- Map2(x, x), or Map3(x, x, x).
	// [Graph.unlink] removes edges by identity and so drops every edge between the
	// pair at once, which means visiting the same parent again would find it
	// already unlinked and still unnecessary, and tear its subgraph down a second
	// time. That is not just wasted work: it re-walks the whole subgraph once per
	// duplicate edge, which compounds to arity^depth over a chain, and it
	// decrements the graph's node count once per repeat.
	//
	// Almost every node has a handful of inputs, where scanning the entries
	// already visited is cheaper than allocating a set. A mapN's arity is
	// unbounded though, so past a threshold the scan would itself be quadratic in
	// the input count.
	if len(parents) > edgeIndexThreshold {
		seen := make(map[Identifier]struct{}, len(parents))
		for _, parent := range parents {
			id := parent.Node().id
			if _, ok := seen[id]; ok {
				continue
			}
			seen[id] = struct{}{}
			graph.removeParent(child, parent)
		}
		return
	}
	for index, parent := range parents {
		if nodeAppearsBefore(parents, index, parent) {
			continue
		}
		graph.removeParent(child, parent)
	}
}

// nodeAppearsBefore reports if node occurs in nodes at an index below limit.
//
// Only used for short lists; see [Graph.removeParents] for the wide case.
func nodeAppearsBefore(nodes []INode, limit int, node INode) bool {
	id := node.Node().id
	for index := 0; index < limit; index++ {
		if nodes[index].Node().id == id {
			return true
		}
	}
	return false
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
	// Tearing a node down is idempotent: removeNode deregisters it, so a node
	// that is no longer registered has already been walked. removeParents skips
	// duplicate edges, which is the case that used to reach here twice, but the
	// guard keeps any other repeated path from re-walking the subgraph and
	// double-decrementing the node count.
	if !parent.Node().inGraph {
		return
	}
	for _, handler := range parent.Node().becameUnnecessaryHandlers() {
		handler()
	}
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
		// unlink both directions: removing only the child from the old parent
		// leaves the old parent behind in the child's parent list, where nothing
		// ever removes it. A bind rewriting its right-hand side goes through here
		// on every rebuild, so a one-sided removal accumulates a stale parent per
		// rebuild -- an unbounded leak of superseded nodes, and a parent list that
		// staleness checks re-walk every pass.
		graph.unlink(child, oldParent)
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
	graph.unlink(child, oldParent)
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
	wasInGraph := node.Node().inGraph
	graph.addNode(node)
	if !wasInGraph {
		for _, handler := range node.Node().becameNecessaryHandlers() {
			handler()
		}
	}
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
	for _, sentinels := range node.Node().nodeSentinels() {
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
	// the inGraph flag answers the "already added" question without hashing the
	// node's 16 byte identifier; bind rebuilds run this per node created, so the
	// second map operation is worth avoiding.
	if gnn.inGraph {
		return
	}
	gnn.inGraph = true
	atomic.AddUint64(&graph.numNodes, 1)
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
	atomic.AddUint64(&graph.numNodes, 1)
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
	atomic.AddUint64(&graph.numNodes, 1)
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
	gn.Node().inGraph = false
	delete(graph.nodes, gn.Node().id)
	graph.nodesMu.Unlock()
	graph.zeroNode(gn)
}

func (graph *Graph) zeroNode(n INode) {
	if n.Node().heightInRecomputeHeap != HeightUnset {
		graph.recomputeHeap.remove(n)
	}

	atomic.AddUint64(&graph.numNodes, ^uint64(0)) // decrement by one

	nn := n.Node()

	// both of these maps are empty outside of stabilization, and tearing down a
	// bind's subgraph calls through here once per node, so check for emptiness
	// before paying to hash the node's identifier.
	graph.handleAfterStabilizationMu.Lock()
	if len(graph.handleAfterStabilization) > 0 {
		delete(graph.handleAfterStabilization, nn.id)
	}
	graph.handleAfterStabilizationMu.Unlock()

	graph.setDuringStabilizationMu.Lock()
	if len(graph.setDuringStabilization) > 0 {
		delete(graph.setDuringStabilization, nn.id)
	}
	graph.setDuringStabilizationMu.Unlock()

	nn.setAt = 0
	nn.changedAt = 0
	nn.recomputedAt = 0

	// mirror how we initialized the node
	nn.valid = true

	nn.parents = nil
	nn.children = nil
	// the slices above may have been backed by storage inside the node, which
	// would otherwise keep the unlinked nodes reachable.
	nn.parentsInline = [1]INode{}
	nn.childrenInline = [1]INode{}
	nn.parentIndex = nil
	nn.childIndex = nil
	nn.observerIndex = nil
	nn.observers = nil

	nn.height = HeightUnset
	nn.heightInRecomputeHeap = HeightUnset
	nn.heightInAdjustHeightsHeap = HeightUnset
}

func (graph *Graph) observeNode(o IObserver, input INode) error {
	graph.addObserver(o)
	wasNecsessary := input.Node().isNecessary()
	input.Node().addObservers(o)
	if !wasNecsessary {
		if err := graph.becameNecessary(input); err != nil {
			return err
		}
	}
	graph.handleAfterStabilizationMu.Lock()
	graph.handleAfterStabilization[o.Node().id] = o.Node().updateHandlers()
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
	// only decorate the context with the stabilization number when a tracer is
	// present; the number is consumed exclusively by trace output, and
	// context.WithValue allocates on every call otherwise.
	if GetTracer(ctx) != nil {
		ctx = WithStabilizationNumber(ctx, graph.stabilizationNum)
		TracePrintln(ctx, "stabilization starting")
	}
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
	// fast path; nothing was set during the pass so there is no work to do
	// and no reason to acquire the lock. This read is safe because all
	// recompute work (the only concurrent writer) has completed by now.
	if len(graph.setDuringStabilization) == 0 {
		return
	}
	graph.setDuringStabilizationMu.Lock()
	defer graph.setDuringStabilizationMu.Unlock()

	if graph.deterministic {
		keys := make([]Identifier, 0, len(graph.setDuringStabilization))
		for key := range graph.setDuringStabilization {
			keys = append(keys, key)
		}
		slices.SortFunc(keys, func(id0, id1 Identifier) int {
			return strings.Compare(id0.String(), id1.String())
		})
		for _, nodeID := range keys {
			n := graph.setDuringStabilization[nodeID]
			_ = n.Node().maybeStabilize(ctx)
			graph.SetStale(n)
		}
	} else {
		for _, n := range graph.setDuringStabilization {
			_ = n.Node().maybeStabilize(ctx)
			graph.SetStale(n)
		}
	}
	clear(graph.setDuringStabilization)
}

func (graph *Graph) stabilizeEndRunUpdateHandlers(ctx context.Context) {
	atomic.StoreInt32(&graph.status, StatusRunningUpdateHandlers)
	// fast path; no update handlers were queued during the pass so we can
	// avoid acquiring the lock and iterating the (empty) map. This read is
	// safe because all recompute work (the only concurrent writer) has
	// completed by the time we reach this point.
	if len(graph.handleAfterStabilization) == 0 {
		return
	}
	graph.handleAfterStabilizationMu.Lock()
	defer graph.handleAfterStabilizationMu.Unlock()

	TracePrintln(ctx, "stabilization calling user update handlers starting")
	defer func() {
		TracePrintln(ctx, "stabilization calling user update handlers complete")
	}()

	if graph.deterministic {
		// we have to sort these so that update handlers fire in order
		keys := make([]Identifier, 0, len(graph.handleAfterStabilization))
		for key := range graph.handleAfterStabilization {
			keys = append(keys, key)
		}
		slices.SortFunc(keys, func(id0, id1 Identifier) int {
			return strings.Compare(id0.String(), id1.String())
		})
		for _, nodeID := range keys {
			for _, uh := range graph.handleAfterStabilization[nodeID] {
				uh(ctx)
			}
		}
	} else {
		for _, updateGroup := range graph.handleAfterStabilization {
			for _, uh := range updateGroup {
				uh(ctx)
			}
		}
	}
	clear(graph.handleAfterStabilization)
}

// recompute starts the recompute cycle for the node
// setting the recomputedAt field and possibly changing the value.
//
// When stabilizing serially, recomputing a node can hand back a single child
// that is safe to recompute immediately; rather than routing that child through
// the recompute heap, this loops on it directly. See recomputeNode for when that
// applies. The loop keeps the chain iterative, so a long run of such nodes costs
// constant stack rather than one frame per node.
func (graph *Graph) recompute(ctx context.Context, n INode, parallel bool) (err error) {
	for n != nil {
		n, err = graph.recomputeNode(ctx, n, parallel)
		if err != nil {
			return
		}
		if n != nil {
			graph.numNodesRecomputedDirectly++
		}
	}
	return
}

// recomputeNode recomputes a single node, returning a child to recompute
// immediately, or nil if the children were queued to the recompute heap.
func (graph *Graph) recomputeNode(ctx context.Context, n INode, parallel bool) (immediate INode, err error) {
	// these counters are only shared when stabilizing in parallel; on the serial
	// path a plain increment is equivalent and avoids a locked instruction per
	// node, which is otherwise a measurable share of a cheap node's recompute.
	if parallel {
		atomic.AddUint64(&graph.numNodesRecomputed, 1)
	} else {
		graph.numNodesRecomputed++
	}

	nn := n.Node()
	nn.numRecomputes++
	nn.recomputedAt = graph.stabilizationNum

	var shouldCutoff bool
	shouldCutoff, err = nn.maybeCutoff(ctx)
	if err != nil {
		for _, eh := range nn.errorHandlers() {
			eh(ctx, err)
		}
		return
	}
	if shouldCutoff {
		return
	}

	if parallel {
		atomic.AddUint64(&graph.numNodesChanged, 1)
	} else {
		graph.numNodesChanged++
	}
	nn.numChanges++

	// a node whose Stabilize mutates graph structure (i.e. a bind swapping out
	// its right-hand-side subgraph) cannot run concurrently with another
	// worker's structural work, so on the parallel path it is serialized under
	// recomputeMu. leaf nodes' Stabilize is left lock-free, which is the whole
	// point of stabilizing in parallel.
	mutatesStructure := parallel && nodeMutatesStructure(n)
	if mutatesStructure {
		graph.recomputeMu.Lock()
	}
	err = nn.maybeStabilize(ctx)
	if mutatesStructure {
		graph.recomputeMu.Unlock()
	}
	if err != nil {
		for _, eh := range nn.errorHandlers() {
			eh(ctx, err)
		}
		return
	}

	nn.changedAt = graph.stabilizationNum
	if handlers := nn.updateHandlers(); len(handlers) > 0 {
		graph.queueUpdateHandlers(parallel, nn.id, handlers)
	}

	if parallel {
		// note we lock recomputeMu rather than the recompute heap's own mutex;
		// it is the lock that also guards bind structural mutation, so that
		// reading children's heights/staleness here is mutually exclusive with
		// a concurrent bind rewriting that same state.
		graph.recomputeMu.Lock()
		for _, c := range nn.children {
			cn := c.Node()
			if cn.childChangedNotifier != nil {
				cn.childChangedNotifier.ChildChanged(n)
			}
			if shouldRecomputeChild(cn, graph.stabilizationNum) {
				graph.recomputeHeap.addNodeUnsafe(c)
			}
		}
		graph.recomputeMu.Unlock()
	} else {
		// hold one child back and queue the rest, then decide whether the held
		// child can be recomputed directly. Deferring that decision until every
		// sibling has been queued is what makes the recompute heap's minimum
		// height sound to test against in canRecomputeImmediately.
		var held INode
		var heldNode *Node
		for _, c := range nn.children {
			cn := c.Node()
			// a node can appear among the children more than once, for instance
			// when it takes the same input twice. It must not be both held back
			// and queued, or it would end up recomputed while still linked into
			// a height block.
			if cn == heldNode {
				continue
			}
			if cn.childChangedNotifier != nil {
				cn.childChangedNotifier.ChildChanged(n)
			}
			if !shouldRecomputeChild(cn, graph.stabilizationNum) {
				continue
			}
			if held != nil {
				graph.recomputeHeap.addNodeUnsafe(held)
			}
			held, heldNode = c, cn
		}
		if held != nil {
			if graph.canRecomputeImmediately(nn, held) {
				immediate = held
			} else {
				graph.recomputeHeap.addNodeUnsafe(held)
			}
		}
	}

	// recompute observers immediately because logically they're
	// children of this node but will not have any children themselves.
	for _, o := range nn.observers {
		if handlers := o.Node().updateHandlers(); len(handlers) > 0 {
			graph.queueUpdateHandlers(parallel, o.Node().id, handlers)
		}
	}
	return
}

// canRecomputeImmediately reports if a child can be recomputed as soon as its
// parent finishes, skipping the trip through the recompute heap.
//
// Adding a node to the recompute heap and taking it back out again is only
// necessary when something else has to be stabilized in between. For a child
// whose only input is the node that just recomputed, nothing else can be
// pending, so the heap round trip is pure overhead. This is the common shape of
// a chain of maps, where it removes the heap from the hot path entirely.
//
// The conditions are:
//
//   - The child takes exactly one input, which is the node that just recomputed,
//     so nothing else can still be pending for it. This is the shape of a chain
//     of maps, and removes the heap from that hot path entirely. Multi-input
//     nodes (map2, mapN, folds) must wait, because a sibling input may still be
//     queued. See the note in _bench/ALGORITHMS.md on why the reference's
//     additional minimum-height bypass is not safe to approximate here.
//
//   - The child's scope has already stabilized, i.e. the parent sits above the
//     height of the bind that created the child. Otherwise the child could still
//     be invalidated by that bind and must not be computed yet.
//
//   - The child is a kind that may be recomputed out of heap order at all; see
//     nodeRequiresHeapOrdering.
//
//   - The child is not an "always" node, because [Graph.Stabilize] re-queues
//     those by inspecting the nodes it pulled from the heap, and a node reached
//     by chaining never appears there.
func (graph *Graph) canRecomputeImmediately(parent *Node, child INode) bool {
	cn := child.Node()
	if cn.always || cn.requiresHeapOrdering || parent.height <= cn.createdIn.scopeHeight() {
		return false
	}
	if len(cn.parents) == 1 {
		return true
	}
	return cn.height <= graph.recomputeHeap.minHeightUnsafe()
}

// nodeRequiresHeapOrdering reports if a node must be taken off the recompute heap
// in height order rather than recomputed directly by the node it depends on.
//
// Bind nodes are excluded because the height relationship the direct-recompute
// checks rely on does not capture their real ordering constraint:
//
//   - A bind-lhs-change node rewrites graph structure, which has to happen in
//     height order relative to the rest of the pass.
//   - A bind main node must not recompute before its own lhs-change node has run,
//     since that is what populates the right-hand side it reads. Its inputs do
//     not express that as an ordinary height relationship, so neither the
//     single-input nor the minimum-height argument establishes it, and running a
//     bind main early can observe a right-hand side that is not yet set.
func nodeRequiresHeapOrdering(n INode) bool {
	switch n.(type) {
	case IBindChange:
		return true
	case IBindMain:
		return true
	default:
		return false
	}
}

// shouldRecomputeChild reports if a child of a just-recomputed node needs to be
// added to the recompute heap, where stabilizationNum is the number of the pass
// currently running.
//
// The checks are ordered cheapest-first and short-circuit: the recompute heap
// membership test is a single field load, whereas isStale can walk the node's
// parents, so testing membership first avoids that walk for the common case of
// a child that is already queued.
//
// The parent scan inside isStale can usually be skipped outright. The caller has
// just stamped the parent's changedAt with the current stabilization number, so
// for a child that has not already been recomputed in this pass, that parent
// alone makes the scan's answer true; the scan would find it and stop. That
// leaves only children carrying their own staleness predicate to ask directly.
func shouldRecomputeChild(cn *Node, stabilizationNum uint64) bool {
	if cn.heightInRecomputeHeap != HeightUnset || !cn.isNecessary() {
		return false
	}
	if !cn.valid {
		return false
	}
	if cn.staler == nil && cn.recomputedAt < stabilizationNum {
		return true
	}
	return cn.isStale()
}

// queueUpdateHandlers records the update handlers to run after stabilization
// for a given node. The handleAfterStabilization map is only contended when
// stabilizing in parallel, so the lock is skipped entirely on the serial path.
func (graph *Graph) queueUpdateHandlers(parallel bool, id Identifier, handlers []func(context.Context)) {
	if parallel {
		graph.handleAfterStabilizationMu.Lock()
		graph.handleAfterStabilization[id] = handlers
		graph.handleAfterStabilizationMu.Unlock()
		return
	}
	graph.handleAfterStabilization[id] = handlers
}

// nodeMutatesStructure reports whether a node's Stabilize can mutate shared
// graph structure (links, heights, node membership) rather than just computing
// a value. Today this is exactly the bind left-hand-side change node, which
// swaps out a bind's right-hand-side subgraph; the invalidation that mutation
// triggers nests within the same Stabilize call. Such nodes must be serialized
// against each other (and against the child-enqueue step) on the parallel path.
func nodeMutatesStructure(n INode) bool {
	_, ok := n.(IBindChange)
	return ok
}
