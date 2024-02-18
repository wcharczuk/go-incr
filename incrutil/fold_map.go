package incrutil

import (
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
)

// FoldMap returns an incremental that takes a map typed incremental as an
// input, an initial value, and a combinator, yielding an incremental
// representing the result of the combinator for the input and zero value.
//
// Usage Note: there is no concept of "left" or "right" with maps in go
// because the iteration order is undefined. As a result, the keys passed
// to the provided function, and their associated values will be assumed
// to be unordered.
func FoldMap[K comparable, V any, O any](
	scope incr.Scope,
	i incr.Incr[map[K]V],
	v0 O,
	fn func(K, V, O) O,
) incr.Incr[O] {
	return incr.WithinScope(scope, &foldMapIncr[K, V, O]{
		n:       incr.NewNode("fold_map"),
		i:       i,
		fn:      fn,
		val:     v0,
		parents: []incr.INode{i},
	})
}

var (
	_ incr.Incr[int]  = (*foldMapIncr[string, float64, int])(nil)
	_ incr.IStabilize = (*foldMapIncr[string, float64, int])(nil)
	_ fmt.Stringer    = (*foldMapIncr[string, float64, int])(nil)
)

type foldMapIncr[K comparable, V any, O any] struct {
	n       *incr.Node
	i       incr.Incr[map[K]V]
	parents []incr.INode
	fn      func(K, V, O) O
	val     O
}

func (fmi *foldMapIncr[K, V, O]) Parents() []incr.INode {
	return fmi.parents
}

func (fmi *foldMapIncr[K, V, O]) String() string { return fmi.n.String() }

func (fmi *foldMapIncr[K, V, O]) Node() *incr.Node { return fmi.n }

func (fmi *foldMapIncr[K, V, O]) Value() O { return fmi.val }

func (fmi *foldMapIncr[K, V, O]) Stabilize(_ context.Context) error {
	new := fmi.i.Value()
	fmi.val = foldMapImpl(new, fmi.val, fmi.fn)
	return nil
}

func foldMapImpl[K comparable, V any, O any](
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
