package incr

import (
	"context"
	"fmt"
)

// Bind lets you swap out an entire subgraph of a computation based
// on a given function and a single input.
//
// A way to think about this, as a sequence:
//
// A given node `a` can be bound to `c` or `d` or more subnodes
// with the value of `a` as the input:
//
//	a -> b.bind() -> c
//
// We might want to, at some point in the future, swap out `c` for `d`
// based on some logic:
//
//	a -> b.bind() -> d
//
// As a result, (a) is a child of (b), and (c) or (d) are children of (b).
// When the bind changes from (c) to (d), (c) is unlinked, and is removed
// as a "child" of (b), preventing it from being considered part of the
// overall computation unless it's referenced by another node in the graph.
//
// More information is available at in the [Janestreet docs].
//
// [Janestreet Docs]: https://github.com/janestreet/incremental/blob/master/src/incremental_intf.ml
func Bind[A, B any](scope Scope, input Incr[A], fn BindFunc[A, B]) BindIncr[B] {
	return BindContext(scope, input, func(_ context.Context, bs Scope, va A) (Incr[B], error) {
		return fn(bs, va), nil
	})
}

// BindFunc is the type of bind function.
type BindFunc[A, B any] func(Scope, A) Incr[B]

// BindContextFunc is the type of bind function.
type BindContextFunc[A, B any] func(context.Context, Scope, A) (Incr[B], error)

// BindContext is like Bind but allows the bind delegate to take a context and return an error.
//
// If an error returned, the bind is aborted, the error listener(s) will fire for the node, and the
// computation will stop.
func BindContext[A, B any](scope Scope, input Incr[A], fn BindContextFunc[A, B]) BindIncr[B] {
	bind := &bind[A, B]{
		graph: scope.scopeGraph(),
		lhs:   input,
		fn:    fn,
	}
	bindLeftChange := &bindLeftChangeIncr[A, B]{
		n:    scope.newNode(KindBindLHSChange),
		bind: bind,
	}
	bindLeftChange.parents[0] = input
	WithinScope(scope, bindLeftChange)
	bind.lhsChange = bindLeftChange
	bindMain := &bindMainIncr[A, B]{
		n:    scope.newNode(KindBind),
		bind: bind,
	}
	bindMain.parentsArray[0] = bindLeftChange
	bindMain.parents = bindMain.parentsArray[:1]
	WithinScope(scope, bindMain)
	bind.main = bindMain

	// propagate errors to main from the left change node
	bindLeftChange.n.extra().onErrorHandlers = append(bindLeftChange.n.extra().onErrorHandlers, func(ctx context.Context, err error) {
		for _, eh := range bindMain.n.errorHandlers() {
			eh(ctx, err)
		}
	})
	// propagate aborted events to main from the left change node
	bindLeftChange.n.extra().onAbortedHandlers = append(bindLeftChange.n.extra().onAbortedHandlers, func(ctx context.Context, err error) {
		for _, eh := range bindMain.n.abortedHandlers() {
			eh(ctx, err)
		}
	})
	return bindMain
}

// BindIncr is a node that implements Bind, which can dynamically swap out
// subgraphs based on input incrementals changing.
//
// BindIncr gives the graph dynamism, but as a result is somewhat expensive to
// compute and should be used tactically.
type BindIncr[A any] interface {
	Incr[A]
	IStabilize
	IShouldBeInvalidated
	IBindMain
	fmt.Stringer
}

// IBindMain holds the methods specific to the bind main node.
type IBindMain interface {
	IParents
	Invalidate()
}

// IBindChange holds the methods specific to the bind-lhs-change node.
type IBindChange interface {
	RightScopeNodes() []INode
}

var (
	_ BindIncr[bool] = (*bindMainIncr[string, bool])(nil)
	_ IStale         = (*bindMainIncr[string, bool])(nil)
	_ Scope          = (*bind[string, bool])(nil)

	_ INode                = (*bindLeftChangeIncr[string, bool])(nil)
	_ IShouldBeInvalidated = (*bindLeftChangeIncr[string, bool])(nil)
	_ IBindChange          = (*bindLeftChangeIncr[string, bool])(nil)
)

// bind is a root struct that holds shared
// information for both the main and the lhs-change.
type bind[A, B any] struct {
	graph    *Graph
	lhs      Incr[A]
	rhs      Incr[B]
	rhsNodes []INode
	// rhsNodesSpare holds the buffer from the rebuild before last, so that
	// successive rebuilds alternate between two buffers rather than allocating a
	// fresh one each time. A rebuild has to keep the previous rebuild's list
	// alive while it invalidates those nodes, so one spare is the minimum needed
	// to reuse capacity at all.
	rhsNodesSpare []INode
	// nodeSlab is where this scope's nodes get their metadata, and nodeSlabSpare holds the
	// slab from the rebuild before last. They alternate for the same reason rhsNodes and
	// rhsNodesSpare do: a rebuild leaves the previous generation intact while it builds the
	// replacement and only tears it down afterwards, so the slots safe to reissue belong to
	// the generation before that one, not the one being replaced.
	nodeSlab      nodeSlab
	nodeSlabSpare nodeSlab
	fn            BindContextFunc[A, B]
	main          *bindMainIncr[A, B]
	lhsChange     *bindLeftChangeIncr[A, B]
}

func (b *bind[A, B]) isTopScope() bool          { return false }
func (b *bind[A, B]) isScopeValid() bool        { return b.main.Node().valid }
func (b *bind[A, B]) isScopeNecessary() bool    { return b.main.Node().isNecessary() }
func (b *bind[A, B]) scopeGraph() *Graph        { return b.graph }
func (b *bind[A, B]) scopeHeight() int          { return b.lhsChange.Node().height }
func (b *bind[A, B]) newIdentifier() Identifier { return b.graph.newIdentifier() }

func (b *bind[A, B]) addScopeNode(n INode) {
	b.rhsNodes = append(b.rhsNodes, n)
}

func (b *bind[A, B]) newNode(kind string) *Node {
	return newNodeIn(&b.nodeSlab, kind)
}

func (b *bind[A, B]) String() string {
	return fmt.Sprintf("{%v}", b.main)
}

type bindMainIncr[A, B any] struct {
	n     *Node
	bind  *bind[A, B]
	value B
	// parents is a slice over parentsArray, which lets the lhs-change node swap
	// the right-hand side in without allocating a new input list on each rebuild.
	parents      []INode
	parentsArray [2]INode
}

func (b *bindMainIncr[A, B]) Parents() (out []INode) {
	return b.parents
}

func (b *bindMainIncr[A, B]) Stale() bool {
	return b.n.recomputedAt == 0 || b.n.isStaleInRespectToParent()
}

func (b *bindMainIncr[A, B]) ShouldBeInvalidated() bool {
	return !b.bind.lhsChange.Node().valid
}

func (b *bindMainIncr[A, B]) Node() *Node { return b.n }

func (b *bindMainIncr[A, B]) Value() (output B) {
	return b.value
}

func (b *bindMainIncr[A, B]) Stabilize(ctx context.Context) error {
	if b.bind.rhs != nil {
		b.value = b.bind.rhs.Value()
	} else {
		var zero B
		b.value = zero
	}
	return nil
}

func (b *bindMainIncr[A, B]) Invalidate() {
	for _, n := range b.bind.rhsNodes {
		GraphForNode(b).invalidateNode(n)
	}
}

func (b *bindMainIncr[A, B]) String() string {
	return b.n.String()
}

type bindLeftChangeIncr[A, B any] struct {
	n    *Node
	bind *bind[A, B]
	// parents is an array rather than a slice so that constructing the node does
	// not allocate a separate input list; [Parents] hands out a slice over it.
	parents [1]INode
}

func (b *bindLeftChangeIncr[A, B]) Parents() []INode {
	return b.parents[:]
}

func (b *bindLeftChangeIncr[A, B]) Node() *Node { return b.n }

func (b *bindLeftChangeIncr[A, B]) ShouldBeInvalidated() bool {
	return !b.bind.lhs.Node().valid
}

func (b *bindLeftChangeIncr[A, B]) RightScopeNodes() []INode {
	return b.bind.rhsNodes
}

func (b *bindLeftChangeIncr[A, B]) Stabilize(ctx context.Context) (err error) {
	oldRightNodes := b.bind.rhsNodes
	oldRhs := b.bind.rhs
	// take the buffer from the rebuild before last; oldRightNodes is still needed
	// below, so the two alternate rather than being reused immediately.
	b.bind.rhsNodes = b.bind.rhsNodesSpare[:0]
	b.bind.rhsNodesSpare = nil
	// Start this generation of nodes in a fresh chunk. Nothing is freed here and no node is
	// disturbed: the nodes just handed out stay where they are, and their chunk stays alive
	// as long as any of them is reachable. Grouping a generation together is what lets the
	// whole generation be collected at once when the right-hand side below replaces it,
	// rather than leaving survivors scattered through a chunk shared with later rebuilds.
	b.bind.nodeSlab, b.bind.nodeSlabSpare = b.bind.nodeSlabSpare, b.bind.nodeSlab
	b.bind.nodeSlab.reset()
	b.bind.rhs, err = b.bind.fn(ctx, b.bind, b.bind.lhs.Value())
	if err != nil {
		return
	}

	main := b.bind.main
	main.parentsArray[0] = b
	if b.bind.rhs != nil {
		main.parentsArray[1] = b.bind.rhs
		main.parents = main.parentsArray[:2]
	} else {
		main.parentsArray[1] = nil
		main.parents = main.parentsArray[:1]
	}

	if err = GraphForNode(b).changeParent(b.bind.main, oldRhs, b.bind.rhs); err != nil {
		return err
	}
	if oldRhs != nil {
		// there is a graph configuration option in js that allows
		// for (2) different behaviors here. the commented out below
		// is if the option is enabled.
		for _, n := range oldRightNodes {
			GraphForNode(b).invalidateNode(n)
		}
		// else {
		// // rescope_nodes_created_on_rhs
		// for _, n := range oldRightNodes {
		// 	n.Node().createdIn = b.bind
		// 	b.bind.addScopeNode(n)
		// }
	}
	GraphForNode(b).propagateInvalidity()
	// oldRightNodes is done being read; keep its capacity for the next rebuild.
	clear(oldRightNodes)
	b.bind.rhsNodesSpare = oldRightNodes
	return nil
}

func (b *bindLeftChangeIncr[A, B]) String() string {
	return b.n.String()
}
