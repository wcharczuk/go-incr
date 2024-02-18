package incrutil

import (
	"context"

	"github.com/wcharczuk/go-incr"
)

// FoldRight folds an array from N to 0 carrying the previous value
// to the next interation, yielding a single value.
func FoldRight[T any, O any](scope incr.Scope, i incr.Incr[[]T], v0 O, fn func(T, O) O) incr.Incr[O] {
	return incr.WithinScope(scope, &foldRightIncr[T, O]{
		n:   incr.NewNode("fold_right"),
		i:   i,
		fn:  fn,
		val: v0,
	})
}

type foldRightIncr[T, O any] struct {
	n   *incr.Node
	i   incr.Incr[[]T]
	fn  func(T, O) O
	val O
}

func (fri *foldRightIncr[T, O]) Parents() []incr.INode {
	return []incr.INode{fri.i}
}

func (fri *foldRightIncr[T, O]) String() string { return fri.n.String() }

func (fri *foldRightIncr[T, O]) Node() *incr.Node { return fri.n }

func (fri *foldRightIncr[T, O]) Value() O { return fri.val }

func (fri *foldRightIncr[T, O]) Stabilize(_ context.Context) error {
	new := fri.i.Value()
	fri.val = foldRight(new, fri.val, fri.fn)
	return nil
}

func foldRight[T, O any](input []T, v0 O, fn func(T, O) O) (o O) {
	o = v0
	for x := len(input) - 1; x >= 0; x-- {
		o = fn(input[x], o)
	}
	return
}
