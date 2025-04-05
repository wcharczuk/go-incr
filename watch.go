package incr

import (
	"context"
	"fmt"
)

// Watch returns a new watch incremental that tracks
// values for a given incremental each time it stabilizes.
func Watch[A any](scope Scope, i Incr[A]) WatchIncr[A] {
	return WithinScope(scope, &watchIncr[A]{
		n:    NewNode(KindWatch),
		incr: i,
	})
}

// WatchIncr is a type that implements the watch interface.
type WatchIncr[A any] interface {
	Incr[A]

	// Reset empties the tracked values.
	Reset()

	// Values returns the input incremental values the [Watch] node
	// has seen through stabilization passes. This array of values will
	// continue to grow until you call [Reset] on the node.
	Values() []A
}

var (
	_ Incr[string] = (*watchIncr[string])(nil)
	_ INode        = (*watchIncr[string])(nil)
	_ IStabilize   = (*watchIncr[string])(nil)
	_ fmt.Stringer = (*watchIncr[string])(nil)
)

type watchIncr[A any] struct {
	n      *Node
	incr   Incr[A]
	value  A
	values []A
}

func (w *watchIncr[A]) Parents() []INode {
	return []INode{w.incr}
}

func (w *watchIncr[A]) Value() A {
	return w.value
}

func (w *watchIncr[A]) Stabilize(ctx context.Context) error {
	w.value = w.incr.Value()
	w.values = append(w.values, w.value)
	return nil
}

func (w *watchIncr[A]) Reset() {
	w.values = nil
}

func (w *watchIncr[A]) Values() []A {
	return w.values
}

func (w *watchIncr[A]) Node() *Node {
	return w.n
}

func (w *watchIncr[A]) String() string { return w.n.String() }
