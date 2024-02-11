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
	IUnobserve
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

func (b *bindIncr[A, B]) Scope() Scope {
	return b.scope
}

func (b *bindIncr[A, B]) didInputChange() bool {
	return b.input.Node().changedAt >= b.n.recomputedAt
}

func (b *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	// did input change?
	//
	// we only want to run the bind fn if the _input_ to this node changes.
	//
	// we do want to propagate changes to the bound node to the bind
	// node's children however, so some trickery is involved.
	if !b.didInputChange() {
		// NOTE (wc): ok so this is a tangle.
		// we halt computation based on boundAt for nodes that
		// set their bound at. So if our bound node triggered
		// this stabilization, and we want the stabilization
		// to continue down to our children, we have to
		// update our boundAt. Event though we didn't bind.
		b.n.boundAt = b.n.graph.stabilizationNum
		return nil
	}

	newIncr, err := b.fn(ctx, b.scope, b.input.Value())
	if err != nil {
		return err
	}
	var bindChanged bool
	if b.bound != nil && newIncr != nil {
		if b.bound.Node().id != newIncr.Node().id {
			bindChanged = true
			b.unlinkOldBound(ctx, b.n.observers...)
			if err := b.linkNewBound(ctx, newIncr); err != nil {
				return err
			}
		}
	} else if newIncr != nil {
		bindChanged = true
		if err := b.linkBindChange(ctx); err != nil {
			return err
		}
		if err := b.linkNewBound(ctx, newIncr); err != nil {
			return err
		}
	} else if b.bound != nil {
		bindChanged = true
		b.unlinkOldBound(ctx, b.n.observers...)
		b.unlinkBindChange(ctx)
	}
	if bindChanged {
		b.n.boundAt = b.n.graph.stabilizationNum
	}
	return nil
}

func (b *bindIncr[A, B]) Unobserve(ctx context.Context, observers ...IObserver) {
	fmt.Printf("unobserve %v\n", b)
	b.unlinkOldBound(ctx, observers...)
	b.unlinkBindChange(ctx)
}

func (b *bindIncr[A, B]) Link(ctx context.Context) (err error) {
	if b.bound != nil {
		err = b.n.graph.adjustHeightsHeap.adjustHeights(b.n.graph.recomputeHeap, b, b.bound)
		if err != nil {
			return
		}
		for _, n := range b.scope.rhsNodes {
			if typed, ok := n.(IBind); ok {
				if err = typed.Link(ctx); err != nil {
					return
				}
			}
		}
	}
	return
}

func (b *bindIncr[A, B]) linkBindChange(ctx context.Context) error {
	b.bindChange = WithinScope(b.n.createdIn, &bindChangeIncr[A, B]{
		n:   NewNode("bind-lhs-change"),
		lhs: b.input,
		rhs: b.bound,
	})
	if b.n.label != "" {
		b.bindChange.n.SetLabel(fmt.Sprintf("%s-change", b.n.label))
	}
	Link(b.bindChange, b.input)
	return b.n.graph.observeSingleNode(b.bindChange, b.n.observers...)
}

func (b *bindIncr[A, B]) unlinkBindChange(ctx context.Context) {
	if b.bindChange != nil {
		if b.bound != nil {
			Unlink(b.bound, b.bindChange)
		}
		Unlink(b.bindChange, b.input)

		// NOTE (wc): we don't do a """typical""" unobserve here because we
		// really don't care; if it's time to unlink our bind change, it's our
		// bind change, there is no way to observe it directly, so we'll just
		// shoot it in the face ourselves.
		b.n.graph.removeNodeFromGraph(b.bindChange)
		b.bindChange = nil
	}
}

func (b *bindIncr[A, B]) linkNewBound(ctx context.Context, newIncr Incr[B]) (err error) {
	b.bound = newIncr
	Link(b.bound, b.bindChange)
	Link(b, b.bound)
	if err = b.n.graph.observeNodes(b.bound, b.n.observers...); err != nil {
		return
	}
	for _, n := range b.scope.rhsNodes {
		if typed, ok := n.(IBind); ok {
			if err = typed.Link(ctx); err != nil {
				return
			}
		}
	}
	TracePrintf(ctx, "%v bound new rhs %v", b, b.bound)
	return
}

func (b *bindIncr[A, B]) unlinkOldBound(ctx context.Context, observers ...IObserver) {
	if b.bound != nil {
		Unlink(b.bound, b.bindChange)
		Unlink(b, b.bound)
		b.n.graph.unobserveNodes(ctx, b.bound, observers...)
		b.removeNodesFromScope(ctx, b.scope, observers...)
		b.bound = nil
		TracePrintf(ctx, "%v unbound old rhs %v", b, b.bound)
	}
}

func (b *bindIncr[A, B]) removeNodesFromScope(ctx context.Context, scope *bindScope, observers ...IObserver) {
	for _, n := range scope.rhsNodes {
		if typed, ok := n.(IUnobserve); ok {
			typed.Unobserve(ctx, observers...)
		}
	}
	scope.rhsNodes = nil
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
	return b.n.String()
}
