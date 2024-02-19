package incr

import (
	"context"
	"fmt"
)

// NewNode returns a new node.
func NewNode(kind string) *Node {
	return &Node{
		id:                        NewIdentifier(),
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
type Node struct {
	// createdIn is the "scope" the node was created in
	createdIn Scope
	// id is a unique identifier for the node
	id Identifier
	// kind is the meta-type of the node
	kind string
	// metadata is any additional metadata a user wants to attach to a node.
	metadata any
	// label is a descriptive string for the
	// node, and is set with `SetLabel`
	label string
	// parents are the nodes that this node depends on, that is
	// parents are nodes that this node takes as inputs
	parents []INode
	// children are the nodes that depend on this node, that is
	// children take this node as an input
	children []INode
	// observers are observer nodes that are attached to this
	// node or its children.
	observers []IObserver
	// valid indicates if the scope that created the node is itself valid
	valid bool
	// forceNecessary forces the necessary state on the node
	forceNecessary bool
	// height is the topological sort pseudo-height of the
	// node and is used to order recomputation
	// it is established when the graph is initialized but
	// can also update if bind nodes change their graphs.
	// largely it represents how many levels of inputs feed into
	// this node, e.g. how many other nodes have to update before
	// this node has to update.
	height int
	// heightInRecomputeHeap is the height of a node in the recompute heap
	heightInRecomputeHeap int
	// heightInAdjustHeightsHeap is the height of a node in the adjust heights heap
	heightInAdjustHeightsHeap int
	// changedAt connotes when the node was changed last,
	// specifically if any of the node's parents were set or bound
	changedAt uint64
	// setAt connotes when the node was set last, specifically
	// for var nodes so that we can track their "changed" state separately
	// from their set state
	setAt uint64
	// recomputedAt connotes when the node was last stabilized
	recomputedAt uint64
	// onUpdateHandlers are functions that are called when the node updates.
	// they are added with `OnUpdate(...)`.
	onUpdateHandlers []func(context.Context)
	// onErrorHandlers are functions that are called when the node updates.
	// they are added with `OnUpdate(...)`.
	onErrorHandlers []func(context.Context, error)
	// stabilizeFn is set during initialization and is a shortcut
	// to the interface sniff for the node for the IStabilize interface.
	stabilizeFn func(context.Context) error
	// shouldBeInvalidated is set during initialization and is a shortcut
	// to the interface sniff for the node for the IStabilize interface.
	shouldBeInvalidatedFn func() bool
	// stale is set during initialization and is a shortcut
	// to the interface sniff for the node for the IStabilize interface.
	staleFn func() bool
	// cutoffFn is set during initialization and is a shortcut
	// to the interface sniff for the node for the ICutoff interface.
	cutoffFn func(context.Context) (bool, error)
	// eachParentFn is a function that nodes can implement to
	// yield their inputs very quickly.
	parentsFn func() []INode
	// invalidateFn is a reference to the nodes invalidate function if present.
	invalidateFn func()
	// observer determines if we treat this as a special necessary state.
	observer bool
	// always determines if we always recompute this node.
	always bool
	// numRecomputes is the number of times we recomputed the node
	numRecomputes uint64
	// numChanges is the number of times we changed the node
	numChanges uint64

	nextInRecomputeHeap     INode
	previousInRecomputeHeap INode
}

//
// Readonly properties
//

// ID returns a unique identifier for the node.
func (n *Node) ID() Identifier {
	return n.id
}

// String returns a string form of the node metadata.
func (n *Node) String() string {
	if n.label != "" {
		return fmt.Sprintf("%s[%s]:%s@%d", n.kind, n.id.Short(), n.label, n.height)
	}
	return fmt.Sprintf("%s[%s]@%d", n.kind, n.id.Short(), n.height)
}

// Set/Get properties

// OnUpdate registers an update handler.
//
// An update handler is called when this node is recomputed.
func (n *Node) OnUpdate(fn func(context.Context)) {
	n.onUpdateHandlers = append(n.onUpdateHandlers, fn)
}

// OnError registers an error handler.
//
// An error handler is called when the stabilize or cutoff
// function for this node returns an error.
func (n *Node) OnError(fn func(context.Context, error)) {
	n.onErrorHandlers = append(n.onErrorHandlers, fn)
}

// Label returns a descriptive label for the node or
// an empty string if one hasn't been provided.
func (n *Node) Label() string {
	return n.label
}

// SetLabel sets the descriptive label on the node.
func (n *Node) SetLabel(label string) {
	n.label = label
}

// Metadata returns user assignable metadata.
func (n *Node) Metadata() any {
	return n.metadata
}

// SetMetadata sets the metadata on the node.
func (n *Node) SetMetadata(md any) {
	n.metadata = md
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
}

func (n *Node) addChildren(children ...INode) {
	n.children = append(n.children, children...)
}

func (n *Node) addParents(parents ...INode) {
	n.parents = append(n.parents, parents...)
}

func (n *Node) addObservers(observers ...IObserver) {
	n.observers = append(n.observers, observers...)
}

func (n *Node) removeChild(id Identifier) {
	n.children = remove(n.children, id)
}

func (n *Node) removeParent(id Identifier) {
	n.parents = remove(n.parents, id)
}

func (n *Node) removeObserver(id Identifier) {
	n.observers = remove(n.observers, id)
}

// maybeCutoff calls the cutoff delegate if it's set, otherwise
// just returns false (effectively _not_ cutting off the computation).
func (n *Node) maybeCutoff(ctx context.Context) (bool, error) {
	if n.cutoffFn != nil {
		return n.cutoffFn(ctx)
	}
	return false, nil
}

func (n *Node) detectCutoff(gn INode) {
	if typed, ok := gn.(ICutoff); ok {
		n.cutoffFn = typed.Cutoff
	}
}

func (n *Node) detectParents(gn INode) {
	if typed, ok := gn.(IParents); ok {
		n.parentsFn = typed.Parents
	}
}

func (n *Node) detectAlways(gn INode) {
	_, n.always = gn.(IAlways)
}

func (n *Node) detectInvalidate(gn INode) {
	if typed, ok := gn.(IBindMain); ok {
		n.invalidateFn = typed.Invalidate
	}
}

func (n *Node) detectObserver(gn INode) {
	_, n.observer = gn.(IObserver)
}

func (n *Node) detectStabilize(gn INode) {
	if typed, ok := gn.(IStabilize); ok {
		n.stabilizeFn = typed.Stabilize
	}
}

func (n *Node) detectStale(gn INode) {
	if typed, ok := gn.(IStale); ok {
		n.staleFn = typed.Stale
	}
}

func (n *Node) detectShouldBeInvalidated(gn INode) {
	if typed, ok := gn.(IShouldBeInvalidated); ok {
		n.shouldBeInvalidatedFn = typed.ShouldBeInvalidated
	}
}

func (n *Node) maybeInvalidate() {
	if n.invalidateFn != nil {
		n.invalidateFn()
	}
}

func (n *Node) maybeStabilize(ctx context.Context) (err error) {
	if n.stabilizeFn != nil {
		if err = n.stabilizeFn(ctx); err != nil {
			return
		}
	}
	return
}

func (n *Node) shouldBeInvalidated() bool {
	if !n.valid {
		return false
	}
	if n.shouldBeInvalidatedFn != nil {
		return n.shouldBeInvalidatedFn()
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
	if n.parentsFn != nil {
		return n.parentsFn()
	}
	return nil
}

func (n *Node) isStaleInRespectToParent() (stale bool) {
	for _, p := range n.nodeParents() {
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
	if n.staleFn != nil {
		return n.staleFn()
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
