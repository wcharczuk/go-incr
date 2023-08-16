package incrutil

import (
	"context"

	"github.com/wcharczuk/go-incr"
)

// DiffSliceByIndicesAdded diffs a slice between stabilizations, yielding an
// incremental that is just the added elements per pass.
func DiffSliceByIndicesAdded[T any](i incr.Incr[[]T]) incr.Incr[[]T] {
	o := &diffSliceByIndicesAddedIncr[T]{
		n: incr.NewNode(),
		i: i,
	}
	incr.Link(o, i)
	return o
}

type diffSliceByIndicesAddedIncr[T any] struct {
	n    *incr.Node
	i    incr.Incr[[]T]
	last int
	val  []T
}

func (dsi *diffSliceByIndicesAddedIncr[T]) String() string {
	return incr.Label(dsi.n, "diff_slice_by_indices_added")
}

func (dsi *diffSliceByIndicesAddedIncr[T]) Node() *incr.Node { return dsi.n }

func (dsi *diffSliceByIndicesAddedIncr[T]) Value() []T { return dsi.val }

func (dsi *diffSliceByIndicesAddedIncr[T]) Stabilize(_ context.Context) error {
	newVal := dsi.i.Value()
	dsi.val, dsi.last = diffSliceByIndicesAdded(dsi.last, newVal)
	return nil
}

func diffSliceByIndicesAdded[T any](previousLast int, value []T) (output []T, last int) {
	if len(value) == 0 {
		return
	}
	last = len(value)
	for x := previousLast; x < len(value); x++ {
		output = append(output, value[x])
	}
	return
}
