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
// More information is available at:
//
//	https://github.com/janestreet/incremental/blob/master/src/incremental_intf.ml
func Bind[A, B any](scope Scope, input Incr[A], fn BindFunc[A, B]) BindIncr[B] {
	return BindContext[A, B](scope, input, func(_ context.Context, bs Scope, va A) (Incr[B], error) {
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
	bindLeftChange := WithinScope(scope, &bindLeftChangeIncr[A, B]{
		n:       NewNode("bind-lhs-change"),
		bind:    bind,
		parents: []INode{input},
	})
	bind.lhsChange = bindLeftChange
	bindMain := WithinScope(scope, &bindMainIncr[A, B]{
		n:       NewNode("bind"),
		bind:    bind,
		parents: []INode{bindLeftChange},
	})
	bind.main = bindMain
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
	graph     *Graph
	lhs       Incr[A]
	rhs       Incr[B]
	rhsNodes  []INode
	fn        BindContextFunc[A, B]
	main      *bindMainIncr[A, B]
	lhsChange *bindLeftChangeIncr[A, B]
}

func (b *bind[A, B]) isTopScope() bool       { return false }
func (b *bind[A, B]) isScopeValid() bool     { return b.main.Node().valid }
func (b *bind[A, B]) isScopeNecessary() bool { return b.main.Node().isNecessary() }
func (b *bind[A, B]) scopeGraph() *Graph     { return b.graph }
func (b *bind[A, B]) scopeHeight() int       { return b.lhs.Node().height }

func (b *bind[A, B]) addScopeNode(n INode) {
	b.rhsNodes = append(b.rhsNodes, n)
}

func (b *bind[A, B]) String() string {
	return fmt.Sprintf("{%v}", b.main)
}

type bindMainIncr[A, B any] struct {
	n       *Node
	bind    *bind[A, B]
	value   B
	parents []INode
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
	n       *Node
	bind    *bind[A, B]
	parents []INode
}

func (b *bindLeftChangeIncr[A, B]) Parents() []INode {
	return b.parents
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
	b.bind.rhsNodes = nil
	b.bind.rhs, err = b.bind.fn(ctx, b.bind, b.bind.lhs.Value())
	if err != nil {
		return
	}

	if b.bind.rhs != nil {
		b.bind.main.parents = []INode{b, b.bind.rhs}
	} else {
		b.bind.main.parents = []INode{b}
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
	return nil
}

func (b *bindLeftChangeIncr[A, B]) String() string {
	return b.n.String()
}
