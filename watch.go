package incr

import (
	"context"
	"fmt"
)

// Watch returns a new watch incremental that tracks
// values for a given incremental.
func Watch[A any](i Incr[A]) WatchIncr[A] {
	o := &watchIncr[A]{
		n:    NewNode(),
		incr: i,
	}
	Link(o, i)
	return o
}

// WatchIncr is a type that implements the watch interface.
type WatchIncr[A any] interface {
	Incr[A]
	Values() []A
	Read() Incr[A]
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

// Value implements Incr[A].
func (w *watchIncr[A]) Value() A {
	return w.value
}

// Read returns an incremental for the watch.
func (w *watchIncr[A]) Read() Incr[A] { return w }

// Stabilize implements Incr[A].
func (w *watchIncr[A]) Stabilize(ctx context.Context) error {
	w.value = w.incr.Value()
	w.values = append(w.values, w.value)
	return nil
}

// Values returns the observed values.
func (w *watchIncr[A]) Values() []A {
	return w.values
}

// Node implements Incr[A].
func (w *watchIncr[A]) Node() *Node {
	return w.n
}

// String implements fmt.Stringer.
func (w *watchIncr[A]) String() string { return FormatNode(w.n, "watch") }
