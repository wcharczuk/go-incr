package incr

import "context"

// Ext allows the user to provide their own incremental
// node type for calculations, which will be integrated with
// the internal metadata of the computation for you.
func Ext[A any](i ExtIncr[A]) Incr[A] {
	ei := &extIncr[A]{
		ei: i,
	}
	ei.n = newNode(ei)
	return ei
}

type extIncr[A any] struct {
	n  *node
	ei ExtIncr[A]
}

func (ei extIncr[A]) Value() A {
	return ei.ei.Value()
}

func (ei extIncr[A]) Stabilize(ctx context.Context) error {
	return ei.ei.Stabilize(ctx)
}

func (ei extIncr[A]) Stale() bool {
	return ei.ei.Stale()
}

func (ei extIncr[A]) getNode() *node {
	return ei.n
}
