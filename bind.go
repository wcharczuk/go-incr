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
func Bind[A, B any](ctx context.Context, input Incr[A], fn func(context.Context, A) Incr[B]) BindIncr[B] {
	return BindContext[A, B](ctx, input, func(ctx context.Context, va A) (Incr[B], error) {
		return fn(ctx, va), nil
	})
}

// BindContext is like Bind but allows the bind delegate to take a context and return an error.
//
// If an error returned, the bind is aborted and the error listener(s) will fire for the node.
func BindContext[A, B any](ctx context.Context, input Incr[A], fn func(context.Context, A) (Incr[B], error)) BindIncr[B] {
	o := &bindIncr[A, B]{
		n:     NewNode(),
		input: input,
		fn:    fn,
		bt:    "bind",
	}
	o.scope = &bindScope{
		lhs:      input,
		bind:     o,
		rhsNodes: newNodeList(),
	}
	Link(o, input)
	return WithinBindScope(ctx, o)
}

// BindIncr is a node that implements Bind, which
// dynamically swaps out entire subgraphs
// based on input incrementals.
type BindIncr[A any] interface {
	Incr[A]
	IStabilize
	IBind
	IUnobserve
	fmt.Stringer
}

var (
	_ BindIncr[bool] = (*bindIncr[string, bool])(nil)
	_ IExpertBind    = (*bindIncr[string, bool])(nil)
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

func (b *bindIncr[A, B]) Bound() INode {
	return b.bound
}

func (b *bindIncr[A, B]) BindChange() INode {
	return b.bindChange
}

func (b *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	newIncr, err := b.fn(withBindScope(ctx, b.scope), b.input.Value())
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

func (b *bindIncr[A, B]) Link(ctx context.Context) (err error) {
	if b.bound != nil {
		children := b.n.Children()
		for _, c := range children {
			if err = link(c, true /*detectCycles*/, b.bound); err != nil {
				return
			}
		}
		for _, n := range b.scope.rhsNodes.list {
			if typed, ok := n.(IBind); ok {
				TracePrintf(ctx, "%v propagating bind link to %v", b, n)
				if err = typed.Link(ctx); err != nil {
					return
				}
			}
		}
		propagateHeightChange(b.bound)
		for _, c := range children {
			propagateHeightChange(c)
		}
	}
	return
}

func (b *bindIncr[A, B]) Unlink(ctx context.Context) {
	b.unlinkOld(ctx)
}

func (b *bindIncr[A, B]) linkBindChange(ctx context.Context) error {
	b.bindChange = &bindChangeIncr[A, B]{
		n:   NewNode(),
		lhs: b.input,
		rhs: b.bound,
	}
	if b.n.label != "" {
		b.bindChange.n.SetLabel(fmt.Sprintf("%s-change", b.n.label))
	}
	if err := link(b.bindChange, true /*detectCycles*/, b.input); err != nil {
		return err
	}
	if err := link(b.bound, true /*detectCycles*/, b.bindChange); err != nil {
		return err
	}
	b.n.graph.observeSingleNode(ctx, b.bindChange, b.n.Observers()...)
	return nil
}

func (b *bindIncr[A, B]) linkNew(ctx context.Context, newIncr Incr[B]) error {
	b.bound = newIncr
	if err := b.linkBindChange(ctx); err != nil {
		return err
	}
	children := b.n.Children()
	for _, c := range children {
		if err := link(c, true /*detectCycles*/, b.bound); err != nil {
			return err
		}
	}
	b.n.graph.observeNodes(ctx, b.bound, b.n.Observers()...)
	for _, n := range b.scope.rhsNodes.list {
		if typed, ok := n.(IBind); ok {
			TracePrintf(ctx, "%v propagating bind link to %v", b, typed)
			if err := typed.Link(ctx); err != nil {
				return err
			}
		}
	}
	propagateHeightChange(b.bound)
	for _, c := range children {
		propagateHeightChange(c)
	}
	TracePrintf(ctx, "%v bound new rhs %v", b, b.bound)
	return nil
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
		b.bound = nil
	}
}

func (b *bindIncr[A, B]) removeNodesFromScope(ctx context.Context, scope *bindScope) {
	rhsNodes := scope.rhsNodes.Values()
	for _, n := range rhsNodes {
		n.Node().createdIn = nil
		if typed, ok := n.(IBind); ok {
			TracePrintf(ctx, "%v propagating bind unlink to %v", b, typed)
			typed.Unlink(ctx)
		}
	}
	scope.rhsNodes.Clear()
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

func (b *bindChangeIncr[A, B]) String() string {
	return b.n.String("bind-lhs-change")
}
