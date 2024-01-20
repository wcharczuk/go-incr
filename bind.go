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
// https://github.com/janestreet/incremental/blob/master/src/incremental_intf.ml#L313C1-L324C52
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
		scope: new(bindScope),
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

type bindScope struct {
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
	if b.n.createdIn != nil {
		b.n.createdIn.nodes.pushUnsafe(newIncr.Node().ID(), newIncr)
	}
	innerBindScope := &bindScope{
		nodes: new(list[Identifier, INode]),
	}
	innerBindScope.nodes.pushUnsafe(newIncr.Node().ID(), newIncr)
	b.bound = bindlhs(b.input, newIncr)
	for _, o := range b.n.Observers() {
		b.n.graph.discoverNodesContext(ctx, o, newIncr)
	}
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

func bindlhs[A, B any](left Incr[A], right Incr[B]) *bindlhsIncr[A, B] {
	o := &bindlhsIncr[A, B]{
		n:     NewNode(),
		left:  left,
		right: right,
	}
	Link(o, left)
	Link(right, o)
	return o
}

type bindlhsIncr[A, B any] struct {
	n     *Node
	left  Incr[A]
	right Incr[B]
}

func (b *bindlhsIncr[A, B]) Node() *Node { return b.n }

func (b *bindlhsIncr[A, B]) Value() (output B) {
	return
}

func (b *bindlhsIncr[A, B]) Stabilize(ctx context.Context) error {
	if err := b.n.createdIn.Stabilize(ctx); err != nil {
		return err
	}
	return nil
}

func (b *bindlhsIncr[A, B]) String() string {
	return b.n.String("bind-lhs-change")
}
