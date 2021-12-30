package incr

import "context"

// Watch returns a new watch incremental that tracks values for a given incremental.
func Watch[A any](i Incr[A]) WatchIncr[A] {
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
type WatchIncr[A any] interface {
	Incr[A]
	Values() []A
}

type watchIncr[A any] struct {
	*node
	incr   Incr[A]
	value  A
	values []A
}

func (w *watchIncr[A]) Value() A {
	return w.value
}

func (w *watchIncr[A]) Stabilize(ctx context.Context) error {
	w.value = w.incr.Value()
	w.values = append(w.values, w.value)
	return nil
}

func (w *watchIncr[A]) Values() []A {
	return w.values
}

func (w *watchIncr[A]) getNode() *node {
	return w.node
}
