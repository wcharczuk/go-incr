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
		scope: &bindScope{
			lhs:      input,
			rhsNodes: newNodeList(),
		},
	}
	Link(o, input)
	return o
}

// BindIncr is a node that implements Bind, which
// dynamically swaps out entire subgraphs
// based on input incrementals.
type BindIncr[A any] interface {
	Incr[A]
	IBind
	IUnobserve
	fmt.Stringer
}

type bindScope struct {
	lhs      INode
	rhsNodes *nodeList
}

var (
	_ Incr[bool]     = (*bindIncr[string, bool])(nil)
	_ BindIncr[bool] = (*bindIncr[string, bool])(nil)
	_ INode          = (*bindIncr[string, bool])(nil)
	_ fmt.Stringer   = (*bindIncr[string, bool])(nil)
)

type bindIncr[A, B any] struct {
	n          *Node
	bt         string
	input      Incr[A]
	fn         func(context.Context, A) (Incr[B], error)
	scope      *bindScope
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

func (b *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	newIncr, err := b.fn(ctx, b.input.Value())
	if err != nil {
		return err
	}
	var bindChanged bool
	if b.bound != nil && newIncr != nil {
		if b.bound.Node().id != newIncr.Node().id {
			bindChanged = true
			b.unlinkOld(ctx)
			b.linkNew(ctx, newIncr)
		}
	} else if newIncr != nil {
		bindChanged = true
		b.linkNew(ctx, newIncr)
	} else if b.bindChange != nil {
		bindChanged = true
		b.unlinkOld(ctx)
	}
	if bindChanged {
		b.n.boundAt = b.n.graph.stabilizationNum
	} else {
		TracePrintf(ctx, "%v unchanged after stabilization", b)
	}
	return nil
}

func (b *bindIncr[A, B]) Unobserve(ctx context.Context) {
	b.unlinkOld(ctx)
}

func (b *bindIncr[A, B]) Link(ctx context.Context) {
	if b.bound != nil {
		children := b.n.Children()
		for _, c := range children {
			Link(c, b.bound)
		}
		for _, c := range children {
			b.n.graph.recomputeHeights(c)
		}
		if typed, ok := b.bound.(IBind); ok {
			typed.Link(ctx)
		}
	}
}

func (b *bindIncr[A, B]) Unlink(ctx context.Context) {
	b.unlinkOld(ctx)
}

func (b *bindIncr[A, B]) linkBindChange(ctx context.Context) {
	b.bindChange = &bindChangeIncr[A, B]{
		n:   NewNode(),
		lhs: b.input,
		rhs: b.bound,
	}
	if b.n.label != "" {
		b.bindChange.n.SetLabel(fmt.Sprintf("%s-change", b.n.label))
	}
	Link(b.bindChange, b.input)
	Link(b.bound, b.bindChange)
	b.n.graph.observeSingleNode(ctx, b.bindChange, b.n.Observers()...)
}

func (b *bindIncr[A, B]) linkNew(ctx context.Context, newIncr Incr[B]) {
	b.bound = newIncr
	b.addNodesToScope(ctx, b.scope, b.bound)

	b.linkBindChange(ctx)
	children := b.n.Children()
	for _, c := range children {
		Link(c, b.bound)
	}
	b.n.graph.observeNodes(ctx, b.bound, b.n.Observers()...)
	for _, c := range children {
		b.n.graph.recomputeHeights(c)
	}
	TracePrintf(ctx, "%v bound new rhs %v", b, b.bound)
}

func (b *bindIncr[A, B]) unlinkBindChange(ctx context.Context) {
	Unlink(b.bindChange, b.input)
	Unlink(b.bound, b.bindChange)
	b.n.graph.unobserveSingleNode(ctx, b.bindChange, b.n.Observers()...)
	b.bindChange = nil
}

func (b *bindIncr[A, B]) unlinkOld(ctx context.Context) {
	if b.bound != nil {
		TracePrintf(ctx, "%v unbinding old rhs %v", b, b.bound)
		b.unlinkBindChange(ctx)
		b.removeNodesFromScope(ctx, b.scope)
		b.n.graph.unobserveNodes(ctx, b.bound, b.n.Observers()...)
		for _, c := range b.n.Children() {
			Unlink(c, b.bound)
		}
		b.bindChange = nil
		b.scope = nil
		b.bound = nil
	}
}

func (b *bindIncr[A, B]) addNodesToScope(ctx context.Context, scope *bindScope, gn INode) {
	gnn := gn.Node()
	if !b.n.graph.IsObserving(gn) {
		TracePrintf(ctx, "%v adding node to scope %v", b, gn)
		scope.rhsNodes.Push(gn)
	}
	parents := gnn.Parents()
	for _, p := range parents {
		b.addNodesToScope(ctx, scope, p)
	}
	if typed, ok := gn.(IBind); ok {
		typed.Link(ctx)
	}
	gnn.createdIn[b.n.id] = scope
}

func (b *bindIncr[A, B]) removeNodesFromScope(ctx context.Context, scope *bindScope) {
	rhsNodes := scope.rhsNodes.Values()
	for _, n := range rhsNodes {
		delete(n.Node().createdIn, b.n.id)
		if len(n.Node().createdIn) == 0 {
			if typed, ok := n.(IBind); ok {
				typed.Unlink(ctx)
			}
		}
	}
}

func (b *bindIncr[A, B]) String() string {
	return b.n.String(b.bt)
}

var (
	_ Incr[bool]   = (*bindChangeIncr[string, bool])(nil)
	_ INode        = (*bindChangeIncr[string, bool])(nil)
	_ fmt.Stringer = (*bindChangeIncr[string, bool])(nil)
)

type bindChangeIncr[A, B any] struct {
	n   *Node
	lhs Incr[A]
	rhs Incr[B]
}

func (b *bindChangeIncr[A, B]) Node() *Node { return b.n }

func (b *bindChangeIncr[A, B]) Value() (output B) {
	if b.rhs != nil {
		output = b.rhs.Value()
	}
	return
}

func (b *bindChangeIncr[A, B]) Stabilize(ctx context.Context) error {
	return nil
}

func (b *bindChangeIncr[A, B]) String() string {
	return b.n.String("bind-lhs-change")
}
