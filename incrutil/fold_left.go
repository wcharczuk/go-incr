package incrutil

import (
	"context"

	"github.com/wcharczuk/go-incr"
)

// FoldLeft folds an array from 0 to N carrying the previous value
// to the next interation, yielding a single value.
func FoldLeft[T any, O any](scope incr.Scope, i incr.Incr[[]T], v0 O, fn func(O, T) O) incr.Incr[O] {
	return incr.WithinScope(scope, &foldLeftIncr[T, O]{
		n:       incr.NewNode("fold_left"),
		i:       i,
		parents: []incr.INode{i},
		fn:      fn,
		val:     v0,
	})
}

type foldLeftIncr[T, O any] struct {
	n       *incr.Node
	i       incr.Incr[[]T]
	parents []incr.INode
	fn      func(O, T) O
	val     O
}

func (fli *foldLeftIncr[T, O]) Parents() []incr.INode {
	return []incr.INode{fli.i}
}

func (fli *foldLeftIncr[T, O]) String() string { return fli.n.String() }

func (fli *foldLeftIncr[T, O]) Node() *incr.Node { return fli.n }

func (fli *foldLeftIncr[T, O]) Value() O { return fli.val }

func (fli *foldLeftIncr[T, O]) Stabilize(_ context.Context) error {
	new := fli.i.Value()
	fli.val = foldLeft(new, fli.val, fli.fn)
	return nil
}

func foldLeft[T, O any](input []T, v0 O, fn func(O, T) O) (o O) {
	o = v0
	for _, v := range input {
		o = fn(o, v)
	}
	return
}
