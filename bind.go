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
func Bind[A, B any](scope Scope, input Incr[A], fn func(Scope, A) Incr[B]) BindIncr[B] {
	return BindContext[A, B](scope, input, func(_ context.Context, bs Scope, va A) (Incr[B], error) {
		return fn(bs, va), nil
	})
}

// BindContext is like Bind but allows the bind delegate to take a context and return an error.
//
// If an error returned, the bind is aborted, the error listener(s) will fire for the node, and the
// computation will stop.
func BindContext[A, B any](scope Scope, input Incr[A], fn func(context.Context, Scope, A) (Incr[B], error)) BindIncr[B] {
	o := WithinScope(scope, &bindIncr[A, B]{
		n:     NewNode("bind"),
		input: input,
		fn:    fn,
	})
	o.scope = &bindScope{
		input: input,
		bind:  o,
	}
	Link(o, input)
	return o
}

// BindIncr is a node that implements Bind, which can dynamically swap out
// subgraphs based on input incrementals changing.
//
// BindIncr gives the graph dynamism, but as a result is somewhat expensive to
// compute and should be used tactically.
type BindIncr[A any] interface {
	Incr[A]
	IStabilize
	IBind
	fmt.Stringer
}

var (
	_ BindIncr[bool] = (*bindIncr[string, bool])(nil)
	_ fmt.Stringer   = (*bindIncr[string, bool])(nil)
)

type bindIncr[A, B any] struct {
	n          *Node
	input      Incr[A]
	fn         func(context.Context, Scope, A) (Incr[B], error)
	rhsScope   *bindScope
	bindChange *bindChangeIncr[A, B]
	bound      Incr[B]
}

func (b *bindIncr[A, B]) Node() *Node { return b.n }

func (b *bindIncr[A, B]) Value() (output B) {
	if b.bound != nil {
		output = b.bound.Value()
	}
	return
}

func (b *bindIncr[A, B]) Bound() INode {
	return b.bound
}

func (b *bindIncr[A, B]) BindChange() INode {
	return b.bindChange
}

func (b *bindIncr[A, B]) Scope() Scope {
	return b.scope
}

func (b *bindIncr[A, B]) didInputChange() bool {
	return b.input.Node().changedAt >= b.n.changedAt
}

func (b *bindIncr[A, B]) Stabilize(ctx context.Context) error {

}

func (b *bindIncr[A, B]) Link(ctx context.Context) (err error) {

}

func (b *bindIncr[A, B]) Invalidate(ctx context.Context) {
	for _, n := range b.scope.rhsNodes {
		b.n.graph.invalidateNode(ctx, n)
	}
}

func (b *bindIncr[A, B]) String() string {
	return b.n.String()
}

var (
	_ Incr[bool]   = (*bindChangeIncr[string, bool])(nil)
	_ INode        = (*bindChangeIncr[string, bool])(nil)
	_ fmt.Stringer = (*bindChangeIncr[string, bool])(nil)
)

type bindChangeIncr[A, B any] struct {
	n     *Node
	lhs   Incr[A]
	rhs   Incr[B]
	scope *bindScope
}

func (b *bindChangeIncr[A, B]) Node() *Node { return b.n }

func (b *bindChangeIncr[A, B]) Stabilize(ctx context.Context) error {
	return nil
}

func (b *bindChangeIncr[A, B]) Invalidate(ctx context.Context) {
	for _, n := range b.scope.rhsNodes {
		b.n.graph.invalidateNode(ctx, n)
	}
}

func (b *bindChangeIncr[A, B]) Value() (output B) {
	if b.rhs != nil {
		output = b.rhs.Value()
	}
	return
}

func (b *bindChangeIncr[A, B]) String() string {
	return b.n.String()
}
