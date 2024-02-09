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
		heightInRecomputeHeap:     -1,
		heightInAdjustHeightsHeap: -1,
		isValid:                   true,
	}
}

// Node is the common metadata for any node in the computation graph.
type Node struct {
	id                        Identifier
	kind                      string
	metadata                  any
	label                     string
	graph                     *Graph
	createdIn                 Scope
	parents                   []INode
	children                  []INode
	observers                 []IObserver
	height                    int
	heightInRecomputeHeap     int
	heightInAdjustHeightsHeap int
	isValid                   bool
	changedAt                 uint64
	setAt                     uint64
	recomputedAt              uint64
	onUpdateHandlers          []func(context.Context)
	onErrorHandlers           []func(context.Context, error)
	onObservedHandlers        []func(IObserver)
	onUnobservedHandlers      []func(IObserver)
	stabilize                 func(context.Context) error
	cutoff                    func(context.Context) (bool, error)
	always                    bool
	numRecomputes             uint64
	numChanges                uint64
	numComputePseudoheights   uint64
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

// OnObserved registers an observed handler.
func (n *Node) OnObserved(fn func(IObserver)) {
	n.onObservedHandlers = append(n.onObservedHandlers, fn)
}

// OnUnobserved registers an unobserved handler.
func (n *Node) OnUnobserved(fn func(IObserver)) {
	n.onUnobservedHandlers = append(n.onUnobservedHandlers, fn)
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

// Parent / Child helpers

// Parents returns the node parent list.
//
// Parents are the nodes that depend on this node, that is
// parents are the nodes that will be recomputed if this node changes.
func (n *Node) Parents() []INode {
	return n.parents
}

// Children returns the node child list.
//
// Children are the nodes that this node depends on, that is
// children are the nodes that if they change will force this
// node to change as well.
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

func (n *Node) isNecessary() bool {
	return len(n.parents) > 0 || len(n.observers) > 0
}

func (n *Node) addChildren(children ...INode) {
	for _, c := range children {
		if !hasKey(n.children, c.Node().id) {
			n.children = append(n.children, c)
		}
	}
}

func (n *Node) addParents(parents ...INode) {
	for _, p := range parents {
		if !hasKey(n.parents, p.Node().id) {
			n.parents = append(n.parents, p)
		}
	}
}

func (n *Node) addObservers(observers ...IObserver) {
	for _, o := range observers {
		if !hasKey(n.observers, o.Node().id) {
			n.observers = append(n.observers, o)
			for _, handler := range n.onObservedHandlers {
				handler(o)
			}
		}
	}
}

func (n *Node) removeChild(id Identifier) {
	n.children = remove(n.children, id)
}

func (n *Node) removeParent(id Identifier) {
	n.parents = remove(n.parents, id)
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

func (n *Node) shouldRecompute() bool {
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
	if n.changedAt > n.recomputedAt {
		return true
	}
	for _, c := range n.children {
		if c.Node().changedAt > n.recomputedAt {
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
