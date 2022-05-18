package incr

import (
	"context"
	"fmt"
)

// NewNode returns a new node.
func NewNode() *Node {
	return &Node{id: NewIdentifier()}
}

// Link is a common helper for setting up node relationships,
// specifically adding a set of "inputs" to a "parent" node.
//
// The reverse of this is `Unlink` on the parent node.
func Link(parent INode, inputs ...INode) {
	parent.Node().addChildren(inputs...)
	for _, gnp := range inputs {
		gnp.Node().addParents(parent)
	}
}

// FormatNode formats a node given a known node type.
func FormatNode(n *Node, nodeType string) string {
	if n.label != "" {
		return fmt.Sprintf("%s[%s]:%s", nodeType, n.id.Short(), n.label)
	}
	return fmt.Sprintf("%s[%s]", nodeType, n.id.Short())
}

// SetStale sets a node as stale.
func SetStale(gn INode) {
	n := gn.Node()
	n.setAt = n.g.stabilizationNum
	n.g.recomputeHeap.Add(gn)
}

// Node is the common metadata for any node in the computation graph.
type Node struct {
	// id is a unique identifier for the node
	id Identifier
	// label is a descriptive string for the
	// node, and is set with `SetLabel`
	label string
	// gs is a shared reference to the graph state
	// for the computation
	g *graph
	// parents are the nodes that depend on this node, that is
	// parents are nodes for which this node is an input
	parents []INode
	// children are the nodes that this node depends on, that is
	// children are inputs to this node
	children []INode
	// height is the topological sort height of the
	// node and is used to order recomputation
	// it is established when the graph is initialized
	// - we may want to switch that to when the node is
	//   strictly added to the graph.
	height int
	// changedAt connotes when the node was changed last,
	// and is a function of the stabilization_num on the
	// graph state
	changedAt uint64
	// setAt connotes when the node was set last, specifically
	// for var nodes so that we can track their changed state separately
	// from their set state, and is a function of the
	// stabilization_num (sn) on the graph state
	setAt uint64
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
	cutoff func(context.Context) bool
	// numRecomputes is the number of times we recomputed the node
	numRecomputes uint64
	// numChanges is the number of times we changed the node
	numChanges uint64
}

// OnUpdate registers an update handler.
func (n *Node) OnUpdate(fn func(context.Context)) {
	n.onUpdateHandlers = append(n.onUpdateHandlers, fn)
}

// OnError registers an error handler.
func (n *Node) OnError(fn func(context.Context, error)) {
	n.onErrorHandlers = append(n.onErrorHandlers, fn)
}

// SetLabel sets the descriptive label on the node.
func (n *Node) SetLabel(label string) {
	n.label = label
}

//
// Internal Helpers
//

// addChildren adds children, or nodes that this node
// depends on, specifically nodes that are inputs to this node.
func (n *Node) addChildren(c ...INode) {
	n.children = append(n.children, c...)
}

// removeChild removes a specific child from the node, specifically
// a node that might have been an input to this node.
func (n *Node) removeChild(id Identifier) {
	var newChildren []INode
	for _, oc := range n.children {
		if oc.Node().id != id {
			newChildren = append(newChildren, oc)
		}
	}
	n.children = newChildren
}

// addParents adds parents, or nodes that depend on this node, specifically
// nodes for which this node is an input.
func (n *Node) addParents(p ...INode) {
	n.parents = append(n.parents, p...)
}

// removeParent removes a parent from the node, specifically
// a node for which this node is an input.
func (n *Node) removeParent(id Identifier) {
	var newParents []INode
	for _, oc := range n.parents {
		if oc.Node().id != id {
			newParents = append(newParents, oc)
		}
	}
	n.parents = newParents
}

// maybeCutoff calls the cutoff delegate if it's set, otherwise
// just returns false (effectively _not_ cutting off the computation).
func (n *Node) maybeCutoff(ctx context.Context) bool {
	if n.cutoff != nil {
		return n.cutoff(ctx)
	}
	return false
}

// detectCutoff detects if a INode (which should be the same
// as as managed by this node reference), implements ICutoff
// and grabs a reference to the Cutoff delegate function.
func (n *Node) detectCutoff(gn INode) {
	if typed, ok := gn.(ICutoff); ok {
		n.cutoff = typed.Cutoff
	}
}

// detectStabilize detects if a INode (which should be the same
// as as managed by this node reference), implements IStabilize
// and grabs a reference to the Stabilize delegate function.
func (n *Node) detectStabilize(gn INode) {
	if typed, ok := gn.(IStabilize); ok {
		n.stabilize = typed.Stabilize
	}
}

// shouldRecompute returns whether or not a given node needs to be recomputed.
func (n *Node) shouldRecompute() bool {
	if n.recomputedAt == 0 {
		return true
	}
	if n.stabilize == nil {
		return false
	}
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

// calculateHeight calculates the nodes height in respect to its children
//
// it is only called during discovery, for subsequent uses the height field
// should be referenced.
func (n *Node) calculateHeight() int {
	var maxChildHeight int
	var childHeight int
	for _, c := range n.children {
		if childHeight = c.Node().calculateHeight(); childHeight > maxChildHeight {
			maxChildHeight = childHeight
		}
	}
	return maxChildHeight + 1
}

// recompute starts the recompute cycle for the node
// setting the recomputedAt field and possibly changing the value.
func (n *Node) recompute(ctx context.Context) error {
	n.g.numNodesRecomputed++
	n.numRecomputes++
	n.recomputedAt = n.g.stabilizationNum
	return n.maybeChangeValue(ctx)
}

// maybeChangeValue checks the cutoff, and calls the stabilization
// delegate if one is set, adding the nodes parents to the recompute heap
// if there are changes.
func (n *Node) maybeChangeValue(ctx context.Context) (err error) {
	if n.maybeCutoff(ctx) {
		return
	}
	n.g.numNodesChanged++
	n.numChanges++
	n.changedAt = n.g.stabilizationNum
	if err = n.maybeStabilize(ctx); err != nil {
		for _, eh := range n.onErrorHandlers {
			eh(ctx, err)
		}
		return
	}
	if len(n.onUpdateHandlers) > 0 {
		n.g.handleAfterStabilization.Push(n.id, n.onUpdateHandlers)
	}
	for _, p := range n.parents {
		if n.g.recomputeHeap.Has(p) {
			continue
		}
		if p.Node().shouldRecompute() {
			n.g.recomputeHeap.Add(p)
		}
	}
	return
}

// maybeStabilize calls the stabilize delegate if one is set.
func (n *Node) maybeStabilize(ctx context.Context) (err error) {
	if n.stabilize != nil {
		if err = n.stabilize(ctx); err != nil {
			return
		}
	}
	return
}
