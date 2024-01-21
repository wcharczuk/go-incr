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
func Bind[A, B any](input Incr[A], fn func(A) Incr[B]) BindIncr[B] {
	return BindContext[A, B](input, func(_ context.Context, va A) (Incr[B], error) {
		return fn(va), nil
	})
}

// BindContext is like Bind but allows the bind delegate to take a context and return an error.
//
// If an error returned, the bind is aborted and the error listener(s) will fire for the node.
func BindContext[A, B any](input Incr[A], fn func(context.Context, A) (Incr[B], error)) BindIncr[B] {
	o := &bindIncr[A, B]{
		n:     NewNode(),
		input: input,
		fn:    fn,
		bt:    "bind",
		scope: newBindScope(input),
	}
	Link(o, input)
	return o
}

// BindIncr is a node that implements Bind, which
// dynamically swaps out entire subgraphs
// based on input incrementals.
type BindIncr[A any] interface {
	Incr[A]
	fmt.Stringer
}

func newBindScope(lhs INode) *bindScope {
	return &bindScope{
		lhs:   lhs,
		nodes: new(list[Identifier, INode]),
	}
}

type bindScope struct {
	lhs   INode
	nodes *list[Identifier, INode]
}

func (bs *bindScope) Stabilize(ctx context.Context) error {
	bs.nodes.Each(func(n INode) {
		n.Node().graph.recomputeHeap.Add(n)
	})
	return nil
}

var (
	_ Incr[bool]     = (*bindIncr[string, bool])(nil)
	_ BindIncr[bool] = (*bindIncr[string, bool])(nil)
	_ INode          = (*bindIncr[string, bool])(nil)
	_ fmt.Stringer   = (*bindIncr[string, bool])(nil)
)

type bindIncr[A, B any] struct {
	n     *Node
	bt    string
	input Incr[A]
	fn    func(context.Context, A) (Incr[B], error)
	scope *bindScope
	bound *bindChangeIncr[A, B]
}

func (b *bindIncr[A, B]) Node() *Node { return b.n }

func (b *bindIncr[A, B]) Value() (output B) {
	if b.bound != nil {
		output = b.bound.Value()
	}
	return
}

func (b *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	// recurse through the bind node's current scope
	// invalidating (but not recomputing!) anything in the current scope.
	if err := b.Bind(ctx); err != nil {
		return err
	}
	if b.n.createdIn != nil {
		b.n.createdIn.nodes.Each(func(n INode) {
			b.n.graph.SetStale(n)
		})
	}
	return nil
}

func (b *bindIncr[A, B]) Bind(ctx context.Context) error {
	newIncr, err := b.fn(ctx, b.input.Value())
	if err != nil {
		return err
	}
	var bindChanged bool
	if b.bound != nil && newIncr != nil {
		if b.bound.Node().id != newIncr.Node().id {
			bindChanged = true
			b.unlinkOld(ctx)
			if err := b.linkNew(ctx, newIncr); err != nil {
				return err
			}
		}
	} else if newIncr != nil {
		bindChanged = true
		if err := b.linkNew(ctx, newIncr); err != nil {
			return err
		}
	} else if b.bound != nil {
		bindChanged = true
		b.unlinkOld(ctx)
	}
	if bindChanged {
		b.n.boundAt = b.n.graph.stabilizationNum
	}
	return nil
}

func (b *bindIncr[A, B]) unlinkOld(ctx context.Context) {

}

func (b *bindIncr[A, B]) linkNew(ctx context.Context, newIncr Incr[B]) error {
	b.bound = bindChange(b.input, newIncr)
	b.scope = &bindScope{
		nodes: new(list[Identifier, INode]),
	}
	b.n.graph.discoverScopeNodesContext(ctx, b.scope, newIncr)
	return nil
}

func (b *bindIncr[A, B]) String() string {
	return b.n.String(b.bt)
}

var (
	_ Incr[bool]     = (*bindIncr[string, bool])(nil)
	_ BindIncr[bool] = (*bindIncr[string, bool])(nil)
	_ INode          = (*bindIncr[string, bool])(nil)
	_ fmt.Stringer   = (*bindIncr[string, bool])(nil)
)

func bindChange[A, B any](left Incr[A], right Incr[B]) *bindChangeIncr[A, B] {
	o := &bindChangeIncr[A, B]{
		n:     NewNode(),
		left:  left,
		right: right,
	}
	Link(o, left)
	Link(right, o)
	return o
}

type bindChangeIncr[A, B any] struct {
	n     *Node
	left  Incr[A]
	right Incr[B]
}

func (b *bindChangeIncr[A, B]) Node() *Node { return b.n }

func (b *bindChangeIncr[A, B]) Value() (output B) {
	return b.left.Value()
}

func (b *bindChangeIncr[A, B]) Stabilize(ctx context.Context) error {
	if err := b.n.createdIn.Stabilize(ctx); err != nil {
		return err
	}
	return nil
}

func (b *bindChangeIncr[A, B]) String() string {
	return b.n.String("bind-lhs-change")
}

/*

In [t >>= f], when [f] is applied to the value of [t], all of the nodes that are created depend on that value ([t]).  If the value of [t] changes, then those nodes no longer make sense because they depend on a stale value.  It would be both wasteful and wrong to recompute any of those "invalid" nodes.

	-> bind is a function that takes a value [t] and passes it to function [f] resulting in a new incremental, which may have many children. as a result, all of those children inherently depend on [t].

So, the implementation maintains the invariant that the height of a necessary node is greater than the height of the left-hand side of the nearest enclosing bind.  That guarantees that stabilization will stabilize the left-hand side before recomputing any nodes created on the right-hand side.

	-> the height of an "observed" node is greater than the height of the left-hand side of the enclosing bind, e.g. the "input" to the bind.

Furthermore, if the left-hand side's value changes, stabilization marks all the nodes on the right-hand side as invalid. Such invalid nodes will typically be unnecessary, but there are pathological cases where they remain necessary.

	-> if the bind input changes, _all_ the children of the incr that is returned by the bind should be marked as stale.

The bind height invariant is accomplished using a special "bind-lhs-change" node, which is a parent of the bind-lhs and a child of the bind result.

	-> bind when it stabilizes creates a "bind-lhs-change" node that is an input to the input to the bind, and takes the bind result as an input (?) this seems insane
		-> i think we would want to invert that, to make the bind-lhsc the child of the lhs and the parent of the bind result, and be a value passthrough.
		-> possible this actually is to make the heights line up such that the rhs is always higher than the lhs by at least (1) extra node?

The incremental state maintains the "current scope", which is the bind whose right-hand side is currently being evaluated, or a special "top" scope if there is no bind in effect. Each node has a [created_in] field set to the scope in effect when the node is created.

	-> when the bind stabilizes, it mints a new scope that is empty, and discovers all the children of the bind result, adding that to the scope.

The implementation keeps for each scope, a singly-linked list of all nodes created in that scope. Invalidation traverses this list, and recurs on bind nodes in it to traverse their scopes as well.

	-> when we set a node stale, if that node is a scope root (i.e. a bind-lhs-change) we should set _that entire scope_ stale.

*/
