package incr

import "context"

// NewNode returns a new node.
func NewNode() *Node {
	return &Node{id: newIdentifier()}
}

// Link is a common helper for setting up nodes.
//
// Specifically it adds a given set of inputs as children
// of a given node, and adds a given node as a parent for
// all the input nodes.
//
// The reverse of this is `Unlink` on the node itself.
func Link(gn GraphNode, inputs ...GraphNode) {
	gn.Node().AddChildren(inputs...)
	for _, gnp := range inputs {
		gnp.Node().AddParents(gn)
	}
}

// Unlink removes a node from the computation graph.
//
// Specifically it removes the given node from the parents
// list of each of its children.
func Unlink(gn GraphNode) {
	id := gn.Node().id
	for _, c := range gn.Node().children {
		c.Node().RemoveParent(id)
	}
	gn.Node().children = nil
}

// Node is the common metadata for any node in the computation graph.
type Node struct {
	// id is a unique identifier for the node.
	id identifier

	// gs is a shared reference to the graph state
	// for the computation
	gs *graphState

	// parents are the nodes that depend on this node
	parents []GraphNode

	// children are the nodes that this node depends on
	children []GraphNode

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

	// stabilize is set during initialization and is a shortcut
	// to the interface sniff for the node for the Stabilizer interface.
	stabilize func(context.Context) error

	// cutoff is set during initialization and is a shortcut
	// to the interface sniff for the node for the Cutoffer interface.
	cutoff func(context.Context) bool

	// numRecomputes is the number of times we recomputed the node
	numRecomputes uint64
}

// AddChildren adds children.
func (n *Node) AddChildren(c ...GraphNode) {
	n.children = append(n.children, c...)
}

// RemoveChild removes a specific child from the node.
func (n *Node) RemoveChild(id identifier) {
	var newChildren []GraphNode
	for _, oc := range n.children {
		if oc.Node().id != id {
			newChildren = append(newChildren, oc)
		}
	}
	n.children = newChildren
}

// AddParents adds parents.
func (n *Node) AddParents(p ...GraphNode) {
	n.parents = append(n.parents, p...)
}

// RemoveParent removes a specific parent from the node.
func (n *Node) RemoveParent(id identifier) {
	var newParents []GraphNode
	for _, oc := range n.parents {
		if oc.Node().id != id {
			newParents = append(newParents, oc)
		}
	}
	n.parents = newParents
}

// OnUpdate registers an update handler.
func (n *Node) OnUpdate(fn func(context.Context)) {
	n.onUpdateHandlers = append(n.onUpdateHandlers, fn)
}

// maybeCutoff calls the cutoff delegate if it's set, otherwise
// just returns false (effectively _not_ cutting off the computation).
func (n *Node) maybeCutoff(ctx context.Context) bool {
	if n.cutoff != nil {
		return n.cutoff(ctx)
	}
	return false
}

// detectCutoff detects if a stabilizer (which should be the same
// as as managed by this node reference), implements Cutoffer
// and grabs a reference to the Cutoff delegate function.
func (n *Node) detectCutoff(gn GraphNode) {
	if typed, ok := gn.(Cutoffer); ok {
		n.cutoff = typed.Cutoff
	}
}

// detectStabilizer detects if a stabilizer (which should be the same
// as as managed by this node reference), implements Cutoffer
// and grabs a reference to the Cutoff delegate function.
func (n *Node) detectStabilizer(gn GraphNode) {
	if typed, ok := gn.(Stabilizer); ok {
		n.stabilize = typed.Stabilize
	}
}

// maybeStabilize calls the stabilize delegate if it's set,
// otherwise is nops.
func (n *Node) maybeStabilize(ctx context.Context) error {
	n.numRecomputes++
	n.recomputedAt = n.gs.sn
	if n.stabilize != nil {
		return n.stabilize(ctx)
	}
	return nil
}

// stale returns whether or not a given node
// needs to be recomputed, specifically
// it will return true if it's uninitialized
// or if any of its "children" or dependent nodes
// are stale themselves.
func (n *Node) stale(ctx context.Context) bool {
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
	var cn *Node
	for _, c := range n.children {
		cn = c.Node()
		if cn.changedAt > n.recomputedAt {
			return true
		}
		if cn.stale(ctx) {
			return true
		}
	}
	return false
}

// calculateHeight calculates the height based on the
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

func (n *Node) recompute(ctx context.Context) error {
	if err := n.maybeStabilize(ctx); err != nil {
		return err
	}
	for _, handler := range n.onUpdateHandlers {
		handler(ctx)
	}
	for _, p := range n.parents {
		if !n.gs.rh.has(p) {
			n.gs.rh.add(p)
		}
	}
	return nil
}
