package incr

import "context"

// DiffSlice diffs a slice between stabilizations, yielding an
// incremental that is just the added elements per pass.
func DiffSlice[T any](i Incr[[]T]) Incr[[]T] {
	o := &diffSliceIncr[T]{
		n: NewNode(),
		i: i,
	}
	Link(o, i)
	return o
}

type diffSliceIncr[T any] struct {
	n    *Node
	i    Incr[[]T]
	last int
	val  []T
}

func (dsi *diffSliceIncr[T]) String() string {
	return FormatNode(dsi.n, "diff_slice")
}

func (dsi *diffSliceIncr[T]) Node() *Node { return dsi.n }

func (dsi *diffSliceIncr[T]) Value() []T { return dsi.val }

func (dsi *diffSliceIncr[T]) Stabilize(_ context.Context) error {
	newVal := dsi.i.Value()
	dsi.val, dsi.last = diffSlice(dsi.last, newVal)
	return nil
}

func diffSlice[T any](previousLast int, value []T) (output []T, last int) {
	if len(value) == 0 {
		return
	}
	last = len(value)
	for x := previousLast; x < len(value); x++ {
		output = append(output, value[x])
	}
	return
}
