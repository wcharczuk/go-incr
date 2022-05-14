package incr

import (
	"context"
	"fmt"
)

// FoldMap returns an incremental that takes a map typed incremental as an
// input, an initial value, and a combinator, yielding an incremental
// representing the result of the combinator for the input and zero value.
//
// Usage Note: there is no concept of "left" or "right" with maps in go
// because the order is pseudo-random. As a result, the key order will
// be what it will be and there is no way to know ahead of time what
// "left" or "right" would even mean in practice.
func FoldMap[K comparable, V any, O any](
	i Incr[map[K]V],
	v0 O,
	fn func(K, V, O) O,
) Incr[O] {
	o := &foldMapIncr[K, V, O]{
		n:   NewNode(),
		i:   i,
		fn:  fn,
		val: v0,
	}
	Link(o, i)
	return o
}

var (
	_ Incr[int]    = (*foldMapIncr[string, float64, int])(nil)
	_ INode        = (*foldMapIncr[string, float64, int])(nil)
	_ IStabilize   = (*foldMapIncr[string, float64, int])(nil)
	_ fmt.Stringer = (*foldMapIncr[string, float64, int])(nil)
)

type foldMapIncr[K comparable, V any, O any] struct {
	n   *Node
	i   Incr[map[K]V]
	fn  func(K, V, O) O
	val O
}

func (fmi *foldMapIncr[K, V, O]) String() string { return FormatNode(fmi.n, "fold_map") }

func (fmi *foldMapIncr[K, V, O]) Node() *Node { return fmi.n }

func (fmi *foldMapIncr[K, V, O]) Value() O { return fmi.val }

func (fmi *foldMapIncr[K, V, O]) Stabilize(_ context.Context) error {
	new := fmi.i.Value()
	fmi.val = foldMap(new, fmi.val, fmi.fn)
	return nil
}

func foldMap[K comparable, V any, O any](
	input map[K]V,
	zero O,
	fn func(K, V, O) O,
) (o O) {
	o = zero
	for k, v := range input {
		o = fn(k, v, o)
	}
	return
}
