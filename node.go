package incr

import (
	"context"
	"fmt"
)

// NewNode returns a new node.
func NewNode(kind string) *Node {
	return &Node{
		kind:                      kind,
		valid:                     true, // start out valid!
		height:                    HeightUnset,
		heightInRecomputeHeap:     HeightUnset,
		heightInAdjustHeightsHeap: HeightUnset,
	}
}

// HeightUnset is a constant that denotes that a height isn't
// strictly set (because heights can be 0, we have to use something
// other than the integer zero value).
const HeightUnset = -1

// Node is the common metadata for any node in the computation graph.
//
// Field order here is deliberate rather than logical: every field that
// stabilization touches per node visit is grouped into the leading bytes of the
// struct, and the fields only used for construction, diagnostics and user
// callbacks follow. The struct spans several cache lines, so interleaving hot
// and cold fields would make each node visit fault in lines that hold nothing
// the hot path needs.
type Node struct {
	//
	// hot fields; read or written on every recompute
	//

	// heightInRecomputeHeap is the height of a node in the recompute heap
	heightInRecomputeHeap int
	// height is the topological sort pseudo-height of the
	// node and is used to order recomputation
	// it is established when the graph is initialized but
	// can also update if bind nodes change their graphs.
	// largely it represents how many levels of inputs feed into
	// this node, e.g. how many other nodes have to update before
	// this node has to update.
	height int
	// recomputedAt connotes when the node was last stabilized
	recomputedAt uint64
	// changedAt connotes when the node was changed last,
	// specifically if any of the node's parents were set or bound
	changedAt uint64
	// setAt connotes when the node was set last, specifically
	// for var nodes so that we can track their "changed" state separately
	// from their set state
	setAt uint64
	// nextInRecomputeHeap and previousInRecomputeHeap are the intrusive links
	// for the node's height block in the recompute heap.
	nextInRecomputeHeap     *Node
	previousInRecomputeHeap *Node
	// parents are the nodes that this node depends on, that is
	// parents are nodes that this node takes as inputs
	parents []INode
	// children are the nodes that depend on this node, that is
	// children take this node as an input
	children []INode
	// parentsInline and childrenInline provide the initial backing array for the
	// two slices above.
	//
	// Most nodes take exactly one input and feed exactly one dependent, so
	// without this both slices allocate on their first append -- together the
	// largest single source of allocations when a bind rebuilds a subgraph.
	// Pointing the slice at storage inside the node removes that allocation
	// while leaving every reader unchanged, since these stay ordinary slices.
	// Growing past one element spills to the heap on its own, because append
	// reallocates when it outgrows the capacity it was handed.
	parentsInline  [1]INode
	childrenInline [1]INode
	// stabilizer, cutoffer and staler are the results of sniffing the node for
	// the corresponding optional interfaces, cached at initialization.
	//
	// These hold the interface value rather than a bound method value
	// (`typed.Stabilize`): capturing a method value allocates a closure per node,
	// and node creation dominates allocation in bind-heavy graphs. Calling
	// through the interface costs the same indirect call.
	stabilizer IStabilize
	cutoffer   ICutoff
	staler     IStale
	// self is the node interface value that owns this metadata.
	//
	// The recompute heap threads its linked lists through *Node rather than
	// INode so that walking a height block is a direct pointer hop instead of
	// an interface method call per node; self is what lets the heap hand the
	// owning INode back to the graph when a node is popped for recomputation.
	// It is set when a node is pushed into the recompute heap, so it is always
	// populated for any node the heap can reach.
	self INode
	// observers are observer nodes that are attached to this
	// node or its children.
	observers []IObserver
	// valid indicates if the scope that created the node is itself valid
	valid bool
	// forceNecessary forces the necessary state on the node
	forceNecessary bool
	// observer determines if we treat this as a special necessary state.
	observer bool
	// always determines if we always recompute this node.
	always bool
	// inGraph tracks whether the node is currently registered with the graph,
	// mirroring its presence in the graph's node map.
	inGraph bool
	// requiresHeapOrdering caches the node-kind test in nodeRequiresHeapOrdering,
	// which is otherwise an interface type switch on every visit to the node as
	// somebody's child.
	requiresHeapOrdering bool
	// graph caches the graph the node was created in, so that [GraphForNode] is a
	// field load rather than a virtual call through the node's scope.
	graph *Graph
	// numRecomputes is the number of times we recomputed the node
	numRecomputes uint64
	// numChanges is the number of times we changed the node
	numChanges uint64
	// id is a unique identifier for the node
	id Identifier

	//
	// cold fields; construction, diagnostics and user callbacks
	//

	// createdIn is the "scope" the node was created in
	createdIn Scope
	// kind is the meta-type of the node
	kind string
	// ext holds the fields most nodes never use; see [nodeExtra].
	ext *nodeExtra
	// heightInAdjustHeightsHeap is the height of a node in the adjust heights heap
	heightInAdjustHeightsHeap int
	// shouldBeInvalidatedProvider, parentsProvider and invalidator are the
	// remaining optional-interface sniffs; see stabilizer above for why these
	// hold interfaces rather than bound method values.
	shouldBeInvalidatedProvider IShouldBeInvalidated
	parentsProvider             IParents
	invalidator                 IBindMain
}

// nodeExtra holds the [Node] fields that most nodes never touch: a user supplied
// label and metadata, sentinel tracking, and the three kinds of user callback.
//
// It is allocated on first write. Keeping these out of [Node] makes every node
// materially smaller, which matters because node allocation is the single largest
// source of garbage in bind-heavy graphs -- each time a bind rewrites its
// right-hand side it allocates a fresh node per node in the new subgraph, and
// those nodes typically use none of these fields.
//
// A nil ext is equivalent to one with every field zero, so readers go through the
// accessors below and only writers call [Node.extra].
type nodeExtra struct {
	metadata          any
	label             string
	sentinels         []ISentinel
	onUpdateHandlers  []func(context.Context)
	onErrorHandlers   []func(context.Context, error)
	onAbortedHandlers []func(context.Context, error)
}

// extra returns the node's auxiliary fields, allocating them if this is the first
// use. Call this only when about to write; readers should use the accessors so
// that reading never allocates.
func (n *Node) extra() *nodeExtra {
	if n.ext == nil {
		n.ext = new(nodeExtra)
	}
	return n.ext
}

func (n *Node) updateHandlers() []func(context.Context) {
	if n.ext == nil {
		return nil
	}
	return n.ext.onUpdateHandlers
}

func (n *Node) errorHandlers() []func(context.Context, error) {
	if n.ext == nil {
		return nil
	}
	return n.ext.onErrorHandlers
}

func (n *Node) abortedHandlers() []func(context.Context, error) {
	if n.ext == nil {
		return nil
	}
	return n.ext.onAbortedHandlers
}

func (n *Node) nodeSentinels() []ISentinel {
	if n.ext == nil {
		return nil
	}
	return n.ext.sentinels
}

//
// Readonly properties
//

// ID returns a unique identifier for the node.
//
// The identifier is set by the scope when you associate the node to a scope
// with [WithinScope], as a result when the node is returned from [NewNode] the
// identifier will be zero until it's associated with a scope.
func (n *Node) ID() Identifier {
	return n.id
}

// String returns a string form of the node metadata.
func (n *Node) String() string {
	if label := n.Label(); label != "" {
		return fmt.Sprintf("%s[%s]:%s@%d", n.kind, n.id.Short(), label, n.height)
	}
	return fmt.Sprintf("%s[%s]@%d", n.kind, n.id.Short(), n.height)
}

// Set/Get properties

// OnUpdate registers an update handler.
//
// An update handler is called when this node is recomputed.
func (n *Node) OnUpdate(fn func(context.Context)) {
	e := n.extra()
	e.onUpdateHandlers = append(e.onUpdateHandlers, fn)
}

// OnError registers an error handler.
//
// An error handler is called when the stabilize or cutoff
// function for this node returns an error.
func (n *Node) OnError(fn func(context.Context, error)) {
	e := n.extra()
	e.onErrorHandlers = append(e.onErrorHandlers, fn)
}

// OnAborted registers an aborted handler.
//
// An aborted handler is called when the stabilize or cutoff
// function for this node is pre-empted by another node erroring.
func (n *Node) OnAborted(fn func(context.Context, error)) {
	e := n.extra()
	e.onAbortedHandlers = append(e.onAbortedHandlers, fn)
}

// Label returns a descriptive label for the node or
// an empty string if one hasn't been provided.
func (n *Node) Label() string {
	if n.ext == nil {
		return ""
	}
	return n.ext.label
}

// SetLabel sets the descriptive label on the node.
func (n *Node) SetLabel(label string) {
	n.extra().label = label
}

// Metadata returns user assignable metadata.
func (n *Node) Metadata() any {
	if n.ext == nil {
		return nil
	}
	return n.ext.metadata
}

// SetMetadata sets the metadata on the node.
func (n *Node) SetMetadata(md any) {
	n.extra().metadata = md
}

// Kind returns the meta type of the node.
func (n *Node) Kind() string {
	return n.kind
}

// SetMetadata sets the metadata on the node.
func (n *Node) SetKind(kind string) {
	n.kind = kind
}

//
// Internal Helpers
//

// initializeFrom detects delegates on the node type.
func (n *Node) initializeFrom(in INode) {
	n.detectAlways(in)
	n.detectCutoff(in)
	n.detectInvalidate(in)
	n.detectObserver(in)
	n.detectParents(in)
	n.detectShouldBeInvalidated(in)
	n.detectStabilize(in)
	n.detectStale(in)
	n.detectRequiresHeapOrdering(in)
}

func (n *Node) detectRequiresHeapOrdering(gn INode) {
	n.requiresHeapOrdering = nodeRequiresHeapOrdering(gn)
}

func (n *Node) addChildren(children ...INode) {
	if n.children == nil {
		n.children = n.childrenInline[:0]
	}
	n.children = append(n.children, children...)
}

func (n *Node) addParents(parents ...INode) {
	if n.parents == nil {
		n.parents = n.parentsInline[:0]
	}
	n.parents = append(n.parents, parents...)
}

func (n *Node) addObservers(observers ...IObserver) {
	n.observers = append(n.observers, observers...)
}

func (n *Node) addSentinels(sentinels ...ISentinel) {
	e := n.extra()
	e.sentinels = append(e.sentinels, sentinels...)
}

func (n *Node) removeChild(id Identifier) {
	n.children, _ = remove(n.children, id)
}

func (n *Node) removeParent(id Identifier) {
	n.parents, _ = remove(n.parents, id)
}

func (n *Node) removeObserver(id Identifier) {
	n.observers, _ = remove(n.observers, id)
}

func (n *Node) removeSentinel(id Identifier) {
	if n.ext == nil {
		return
	}
	n.ext.sentinels, _ = remove(n.ext.sentinels, id)
}

// maybeCutoff calls the cutoff delegate if it's set, otherwise
// just returns false (effectively _not_ cutting off the computation).
func (n *Node) maybeCutoff(ctx context.Context) (bool, error) {
	if n.cutoffer != nil {
		return n.cutoffer.Cutoff(ctx)
	}
	return false, nil
}

func (n *Node) detectCutoff(gn INode) {
	if typed, ok := gn.(ICutoff); ok {
		n.cutoffer = typed
	}
}

func (n *Node) detectParents(gn INode) {
	if typed, ok := gn.(IParents); ok {
		n.parentsProvider = typed
	}
}

func (n *Node) detectAlways(gn INode) {
	_, n.always = gn.(IAlways)
}

func (n *Node) detectInvalidate(gn INode) {
	if typed, ok := gn.(IBindMain); ok {
		n.invalidator = typed
	}
}

func (n *Node) detectObserver(gn INode) {
	_, n.observer = gn.(IObserver)
}

func (n *Node) detectStabilize(gn INode) {
	if typed, ok := gn.(IStabilize); ok {
		n.stabilizer = typed
	}
}

func (n *Node) detectStale(gn INode) {
	if typed, ok := gn.(IStale); ok {
		n.staler = typed
	}
}

func (n *Node) detectShouldBeInvalidated(gn INode) {
	if typed, ok := gn.(IShouldBeInvalidated); ok {
		n.shouldBeInvalidatedProvider = typed
	}
}

func (n *Node) maybeInvalidate() {
	if n.invalidator != nil {
		n.invalidator.Invalidate()
	}
}

func (n *Node) maybeStabilize(ctx context.Context) (err error) {
	if n.stabilizer != nil {
		if err = n.stabilizer.Stabilize(ctx); err != nil {
			return
		}
	}
	return
}

func (n *Node) shouldBeInvalidated() bool {
	if !n.valid {
		return false
	}
	if n.shouldBeInvalidatedProvider != nil {
		return n.shouldBeInvalidatedProvider.ShouldBeInvalidated()
	}
	// s/has_invalid_child/has_invalid_parent/g
	for _, p := range n.parents {
		if !p.Node().valid {
			return true
		}
	}
	return false
}

func (n *Node) nodeParents() []INode {
	if n.parentsProvider != nil {
		return n.parentsProvider.Parents()
	}
	return nil
}

func (n *Node) isStaleInRespectToParent() (stale bool) {
	for _, p := range n.parents {
		if p.Node().changedAt > n.recomputedAt {
			stale = true
			return
		}
	}
	return
}

func (n *Node) isStale() bool {
	if !n.valid {
		return false
	}
	if n.staler != nil {
		return n.staler.Stale()
	}
	return n.recomputedAt == 0 || n.isStaleInRespectToParent()
}

func (n *Node) isNecessary() bool {
	if n.observer {
		return true
	}
	if n.forceNecessary {
		return true
	}
	return len(n.children) > 0 || len(n.observers) > 0
}
