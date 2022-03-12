package incr

import "context"

// Watch returns a new watch incremental that tracks values for a given incremental.
func Watch[A comparable](i Incr[A]) WatchIncr[A] {
	w := &watchIncr[A]{
		incr: i,
	}
	w.node = newNode(
		w,
		optNodeChildOf(i),
	)
	return w
}

// WatchIncr extends incr to include Values.
type WatchIncr[A comparable] interface {
	Incr[A]
	Values() []A
}

type watchIncr[A comparable] struct {
	*node
	incr   Incr[A]
	value  A
	values []A
}

func (w *watchIncr[A]) Value() A {
	return w.value
}

func (w *watchIncr[A]) Stabilize(ctx context.Context) error {
	newValue := w.incr.Value()
	if w.value != newValue {
		w.value = newValue
		w.values = append(w.values, w.value)
	}
	return nil
}

func (w *watchIncr[A]) Stale() bool { return false }

func (w *watchIncr[A]) Values() []A {
	return w.values
}

func (v *watchIncr[A]) getValue() any { return v.Value() }

func (w *watchIncr[A]) getNode() *node {
	return w.node
}
