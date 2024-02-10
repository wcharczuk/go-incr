package incr

import "context"

// FoldLeft folds an array from 0 to N carrying the previous value
// to the next interation, yielding a single value.
func FoldLeft[T, O any](scope *BindScope, i Incr[[]T], v0 O, fn func(O, T) O) Incr[O] {
	o := &foldLeftIncr[T, O]{
		n:   NewNode("fold_left"),
		i:   i,
		fn:  fn,
		val: v0,
	}
	Link(o, i)
	return WithinBindScope(scope, o)
}

type foldLeftIncr[T, O any] struct {
	n   *Node
	i   Incr[[]T]
	fn  func(O, T) O
	val O
}

func (fli *foldLeftIncr[T, O]) String() string { return fli.n.String() }

func (fli *foldLeftIncr[T, O]) Node() *Node { return fli.n }

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
