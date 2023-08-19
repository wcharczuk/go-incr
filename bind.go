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
func Bind[A, B any](a Incr[A], fn func(A) Incr[B]) BindIncr[B] {
	return BindContext[A, B](a, func(_ context.Context, va A) (Incr[B], error) {
		return fn(va), nil
	})
}

// BindContext is like Bind but allows the bind delegate to take a context and return an error.
func BindContext[A, B any](a Incr[A], fn func(context.Context, A) (Incr[B], error)) BindIncr[B] {
	o := &bindIncr[A, B]{
		n:        NewNode(),
		a:        a,
		fn:       fn,
		bindType: "bind",
	}
	Link(o, a)
	return o
}

// BindIncr is a node that implements Bind, which
// dynamically swaps out entire subgraphs
// based on input incrementals.
type BindIncr[A any] interface {
	Incr[A]
}

var (
	_ Incr[bool]     = (*bindIncr[string, bool])(nil)
	_ BindIncr[bool] = (*bindIncr[string, bool])(nil)
	_ INode          = (*bindIncr[string, bool])(nil)
	_ IStabilize     = (*bindIncr[string, bool])(nil)
	_ fmt.Stringer   = (*bindIncr[string, bool])(nil)
)

type bindIncr[A, B any] struct {
	n        *Node
	a        Incr[A]
	fn       func(context.Context, A) (Incr[B], error)
	bound    Incr[B]
	bindType string
	value    B
}

func (b *bindIncr[A, B]) Node() *Node { return b.n }

func (b *bindIncr[A, B]) Value() B { return b.value }

func (b *bindIncr[A, B]) Stabilize(ctx context.Context) error {
	oldIncr := b.bound
	newIncr, err := b.fn(ctx, b.a.Value())
	if err != nil {
		return err
	}

	if oldIncr == nil {
		Link(newIncr, b)
		b.Node().graph.discoverAllNodes(newIncr)
		newIncr.Node().changedAt = b.Node().graph.stabilizationNum
		if err := newIncr.Node().maybeStabilize(ctx); err != nil {
			return err
		}
		b.bound = newIncr
		b.value = b.bound.Value()
		return nil
	}

	if oldIncr.Node().id == newIncr.Node().id {
		return nil
	}

	// "unlink" the old node from the bind node
	b.Node().parents = filterNodes(b.Node().parents, func(p INode) bool {
		return oldIncr.Node().id != p.Node().id
	})
	oldIncr.Node().children = filterNodes(oldIncr.Node().children, func(c INode) bool {
		return b.Node().id != c.Node().id
	})
	b.Node().graph.undiscoverAllNodes(oldIncr)

	// link the new value as the parent
	// of the bind node, specifically
	// that b is an input to newValue
	Link(newIncr, b)
	b.Node().graph.discoverAllNodes(newIncr)
	newIncr.Node().changedAt = b.Node().graph.stabilizationNum
	if err := newIncr.Node().maybeStabilize(ctx); err != nil {
		return err
	}

	b.bound = newIncr
	b.value = b.bound.Value()
	return nil
}

func (b *bindIncr[A, B]) String() string {
	return b.n.String(b.bindType)
}

func filterNodes(nodes []INode, filter func(INode) bool) (out []INode) {
	for _, n := range nodes {
		if filter(n) {
			out = append(out, n)
		}
	}
	return out
}
