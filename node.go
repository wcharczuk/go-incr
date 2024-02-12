package incr

import (
	"context"
	"fmt"
	"sync"
)

// NewNode returns a new node.
func NewNode(kind string) *Node {
	return &Node{
		id:             NewIdentifier(),
		kind:           kind,
		parentLookup:   make(set[Identifier]),
		childLookup:    make(set[Identifier]),
		observerLookup: make(set[Identifier]),
	}
}

// Node is the common metadata for any node in the computation graph.
type Node struct {
	// id is a unique identifier for the node
	id Identifier
	// kind is the meta-type of the node
	kind string
	// metadata is any additional metadata a user wants to attach to a node.
	metadata any
	// graph is the graph this node is attached to currently.
	graph *Graph
	// label is a descriptive string for the
	// node, and is set with `SetLabel`
	label string
	// parents are the nodes that this node depends on, that is
	// parents are nodes that this node takes as inputs
	parents        []INode
	parentLookupMu sync.Mutex
	parentLookup   set[Identifier]
	// children are the nodes that depend on this node, that is
	// children take this node as an input
	children      []INode
	childLookupMu sync.Mutex
	childLookup   set[Identifier]
	// observers are observer nodes that are attached to this
	// node or its children.
	observers        []IObserver
	observerLookupMu sync.Mutex
	observerLookup   set[Identifier]
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
	// boundAt connotes when the node was bound last, specifically
	// for bind nodes so that we can track their changed state separately
	// from their bound state
	boundAt uint64
	// recomputedAt connotes when the node was last stabilized
	recomputedAt uint64
	// onUpdateHandlers are functions that are called when the node updates.
	// they are added with `OnUpdate(...)`.
	onUpdateHandlers []func(context.Context)
	// onErrorHandlers are functions that are called when the node updates.
	// they are added with `OnUpdate(...)`.
	onErrorHandlers []func(context.Context, error)
	// stabilize is set during initialization and is a shortcut
	// to the interface sniff for the node for the IStabilize interface.
	stabilize func(context.Context) error
	// cutoff is set during initialization and is a shortcut
	// to the interface sniff for the node for the ICutoff interface.
	cutoff func(context.Context) (bool, error)
	// always determines if we always recompute this node.
	always bool
	// numRecomputes is the number of times we recomputed the node
	numRecomputes uint64
	// numChanges is the number of times we changed the node
	numChanges uint64
	// createdIn is the "scope" the node was created in
	createdIn Scope
	// numComputePseudoHeights is the number of times we call "computePseudoheight"
	// on this particular node
	numComputePseudoHeights uint64
	// numRecomputeHeights is incremented everytime we recompute a given node's height
	numRecomputeHeights uint64
}

func nodeSorter(a, b INode) int {
	if a.Node().height == b.Node().height {
		aID := a.Node().ID().String()
		bID := b.Node().ID().String()
		if aID == bID {
			return 0
		} else if aID > bID {
			return -1
		}
		return 1
	} else if a.Node().height > b.Node().height {
		return -1
	}
	return 1
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
func (n *Node) OnUpdate(fn func(context.Context)) {
	n.onUpdateHandlers = append(n.onUpdateHandlers, fn)
}

// OnError registers an error handler.
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

// Parent / Child helpers

// Parents returns the node parent list.
func (n *Node) Parents() []INode {
	return n.parents
}

// Parents returns the node child list.
func (n *Node) Children() []INode {
	return n.children
}

// Observers returns the node observer list.
func (n *Node) Observers() []IObserver {
	return n.observers
}

//
// Internal Helpers
//

func (n *Node) addChildren(children ...INode) {
	n.childLookupMu.Lock()
	defer n.childLookupMu.Unlock()

	for _, c := range children {
		if !n.childLookup.has(c.Node().id) {
			n.children = append(n.children, c)
			n.childLookup.add(c.Node().id)
		}
	}
}

func (n *Node) addParents(parents ...INode) {
	n.parentLookupMu.Lock()
	defer n.parentLookupMu.Unlock()

	for _, p := range parents {
		if !n.parentLookup.has(p.Node().id) {
			n.parents = append(n.parents, p)
			n.parentLookup.add(p.Node().id)
		}
	}
}

func (n *Node) addObservers(observers ...IObserver) {
	n.observerLookupMu.Lock()
	defer n.observerLookupMu.Unlock()
	for _, o := range observers {
		if !n.observerLookup.has(o.Node().id) {
			n.observers = append(n.observers, o)
			n.observerLookup.add(o.Node().id)
		}
	}
}

func (n *Node) hasChild(in INode) (ok bool) {
	n.childLookupMu.Lock()
	_, ok = n.childLookup[in.Node().id]
	n.childLookupMu.Unlock()
	return
}

func (n *Node) hasParent(in INode) (ok bool) {
	n.parentLookupMu.Lock()
	_, ok = n.parentLookup[in.Node().id]
	n.parentLookupMu.Unlock()
	return
}

func (n *Node) hasObserver(o IObserver) (ok bool) {
	n.observerLookupMu.Lock()
	_, ok = n.observerLookup[o.Node().id]
	n.observerLookupMu.Unlock()
	return
}

func (n *Node) removeChild(id Identifier) {
	n.childLookupMu.Lock()
	n.children = remove(n.children, id)
	delete(n.childLookup, id)
	n.childLookupMu.Unlock()
}

func (n *Node) removeParent(id Identifier) {
	n.parentLookupMu.Lock()
	n.parents = remove(n.parents, id)
	delete(n.parentLookup, id)
	n.parentLookupMu.Unlock()
}

func (n *Node) removeObserver(id Identifier) {
	n.observerLookupMu.Lock()
	n.observers = remove(n.observers, id)
	delete(n.observerLookup, id)
	n.observerLookupMu.Unlock()
}

// maybeCutoff calls the cutoff delegate if it's set, otherwise
// just returns false (effectively _not_ cutting off the computation).
func (n *Node) maybeCutoff(ctx context.Context) (bool, error) {
	if n.cutoff != nil {
		return n.cutoff(ctx)
	}
	return false, nil
}

// detectCutoff detects if a INode (which should be the same
// as as managed by this node reference), implements ICutoff
// and grabs a reference to the Cutoff delegate function.
func (n *Node) detectCutoff(gn INode) {
	if typed, ok := gn.(ICutoff); ok {
		n.cutoff = typed.Cutoff
	}
}

// detectAlways detects if a INode (which should be the same
// as as managed by this node reference), implements IAlways.
func (n *Node) detectAlways(gn INode) {
	_, n.always = gn.(IAlways)
}

// detectStabilize detects if a INode (which should be the same
// as as managed by this node reference), implements IStabilize
// and grabs a reference to the Stabilize delegate function.
func (n *Node) detectStabilize(gn INode) {
	if typed, ok := gn.(IStabilize); ok {
		n.stabilize = typed.Stabilize
	}
}

// ShouldRecompute returns whether or not a given node needs to be recomputed.
func (n *Node) ShouldRecompute() bool {
	// we should always recompute on the first stabilization
	if n.recomputedAt == 0 {
		return true
	}
	if n.always {
		return true
	}

	// if a node can't stabilize, return false
	if n.stabilize == nil {
		return false
	}

	// if the node was marked stale explicitly
	// either because it is a var or was
	// called as a parameter to `graph.SetStale`
	if n.setAt > n.recomputedAt {
		return true
	}
	// if the node had a bind change recently
	if n.boundAt > n.recomputedAt {
		return true
	}
	if n.changedAt > n.recomputedAt {
		return true
	}
	for _, p := range n.parents {
		// NOTE (wc): we treat nodes that set "boundAt", i.e. bind nodes
		// specially. they really only need to propagate their changes
		// to children if the node they're bound to changes, or if
		// the node they're bound to's value changes.
		if p.Node().boundAt > 0 {
			if p.Node().boundAt > n.recomputedAt {
				return true
			}
			// don't consider changed at for bind nodes at all!
			continue
		}
		if p.Node().changedAt > n.recomputedAt {
			return true
		}
	}
	return false
}

func (n *Node) maybeStabilize(ctx context.Context) (err error) {
	if n.stabilize != nil {
		if err = n.stabilize(ctx); err != nil {
			return
		}
	}
	return
}
