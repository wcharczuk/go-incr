package incr

import (
	"context"
	"fmt"
)

// Bind lets you swap out an entire subgraph of a computation based
// on a given function and two inputs.
func Bind2[A, B, C any](scope *BindScope, inputA Incr[A], inputB Incr[B], fn func(*BindScope, A, B) Incr[C]) BindIncr[C] {
	return Bind2Context[A, B, C](scope, inputA, inputB, func(_ context.Context, bs *BindScope, va A, vb B) (Incr[C], error) {
		return fn(bs, va, vb), nil
	})
}

// Bind2Context is like Bind2 but allows the bind delegate to take a context and return an error.
//
// If an error returned, the bind is aborted and the error listener(s) will fire for the node.
func Bind2Context[A, B, C any](scope *BindScope, inputA Incr[A], inputB Incr[B], fn func(context.Context, *BindScope, A, B) (Incr[C], error)) Bind2Incr[C] {
	o := &bind2Incr[A, B, C]{
		n:      NewNode(),
		inputA: inputA,
		inputB: inputB,
		fn:     fn,
	}
	o.scope = &BindScope{
		bind:     o,
		rhsNodes: newNodeList(),
	}
	Link(o, inputA)
	Link(o, inputB)
	return WithinBindScope(scope, o)
}

// Bind2Incr is a node that implements Bind, which
// dynamically swaps out entire subgraphs
// based on input incrementals.
type Bind2Incr[A any] interface {
	Incr[A]
	IStabilize
	IBind
	IUnobserve
	fmt.Stringer
}

type bind2Incr[A, B, C any] struct {
	n          *Node
	inputA     Incr[A]
	inputB     Incr[B]
	fn         func(context.Context, *BindScope, A, B) (Incr[C], error)
	bound      Incr[C]
	bindChange *bindChange2Incr[A, B, C]
	scope      *BindScope
}

func (b *bind2Incr[A, B, C]) Node() *Node { return b.n }

func (b *bind2Incr[A, B, C]) Value() (output C) {
	if b.bound != nil {
		output = b.bound.Value()
	}
	return
}

func (b *bind2Incr[A, B, C]) Bound() INode {
	return b.bound
}

func (b *bind2Incr[A, B, C]) BindChange() INode {
	return b.bindChange
}

func (b *bind2Incr[A, B, C]) Stabilize(ctx context.Context) error {
	newIncr, err := b.fn(ctx, b.scope, b.inputA.Value(), b.inputB.Value())
	if err != nil {
		return err
	}
	var bindChanged bool
	if b.bound != nil && newIncr != nil {
		if b.bound.Node().id != newIncr.Node().id {
			bindChanged = true
			b.unlinkOld(ctx, b.n.Observers()...)
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
		b.unlinkOld(ctx, b.n.Observers()...)
	}
	if bindChanged {
		b.n.boundAt = b.n.graph.stabilizationNum
	} else {
		TracePrintf(ctx, "%v unchanged after stabilization", b)
	}
	return nil
}

func (b *bind2Incr[A, B, C]) Unobserve(ctx context.Context, observers ...IObserver) {
	b.unlinkOld(ctx, observers...)
}

func (b *bind2Incr[A, B, C]) Link(ctx context.Context) (err error) {
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

func (b *bind2Incr[A, B, C]) linkBindChange(ctx context.Context) error {
	b.bindChange = &bindChange2Incr[A, B, C]{
		n:    NewNode(),
		lhsA: b.inputA,
		lhsB: b.inputB,
		rhs:  b.bound,
	}
	if b.n.label != "" {
		b.bindChange.n.SetLabel(fmt.Sprintf("%s-change", b.n.label))
	}
	if err := link(b.bindChange, true /*detectCycles*/, b.inputA); err != nil {
		return err
	}
	if err := link(b.bindChange, true /*detectCycles*/, b.inputB); err != nil {
		return err
	}
	if err := link(b.bound, true /*detectCycles*/, b.bindChange); err != nil {
		return err
	}
	b.n.graph.observeSingleNode(b.bindChange, b.n.Observers()...)
	return nil
}

func (b *bind2Incr[A, B, C]) linkNew(ctx context.Context, newIncr Incr[C]) error {
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
	b.n.graph.observeNodes(b.bound, b.n.Observers()...)
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

func (b *bind2Incr[A, B, C]) unlinkBindChange(ctx context.Context) {
	Unlink(b.bindChange, b.inputA)
	Unlink(b.bindChange, b.inputB)
	Unlink(b.bound, b.bindChange)
}

func (b *bind2Incr[A, B, C]) unlinkOld(ctx context.Context, observers ...IObserver) {
	if b.bound != nil {
		TracePrintf(ctx, "%v unbinding old rhs %v", b, b.bound)
		b.unlinkBindChange(ctx)
		b.removeNodesFromScope(ctx, b.scope)
		b.n.graph.unobserveNodes(ctx, b.bound, observers...)
		for _, c := range b.n.Children() {
			Unlink(c, b.bound)
		}
		b.n.graph.unobserveSingleNode(ctx, b.bindChange, observers...)
		b.bindChange = nil
		b.bound = nil
	}
}

func (b *bind2Incr[A, B, C]) removeNodesFromScope(ctx context.Context, scope *BindScope) {
	rhsNodes := scope.rhsNodes.Values()
	for _, n := range rhsNodes {
		n.Node().createdIn = nil
	}
	scope.rhsNodes.Clear()
}

func (b *bind2Incr[A, B, C]) String() string {
	return b.n.String("bind2")
}

var (
	_ Incr[bool]   = (*bindChangeIncr[string, bool])(nil)
	_ INode        = (*bindChangeIncr[string, bool])(nil)
	_ fmt.Stringer = (*bindChangeIncr[string, bool])(nil)
)

type bindChange2Incr[A, B, C any] struct {
	n    *Node
	lhsA Incr[A]
	lhsB Incr[B]
	rhs  Incr[C]
}

func (b *bindChange2Incr[A, B, C]) Node() *Node { return b.n }

func (b *bindChange2Incr[A, B, C]) Value() (output C) {
	if b.rhs != nil {
		output = b.rhs.Value()
	}
	return
}

func (b *bindChange2Incr[A, B, C]) String() string {
	return b.n.String("bind-2-lhs-change")
}
