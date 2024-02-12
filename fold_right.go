package incr

import "context"

// FoldRight folds an array from N to 0 carrying the previous value
// to the next interation, yielding a single value.
func FoldRight[T, O any](scope Scope, i Incr[[]T], v0 O, fn func(T, O) O) Incr[O] {
	o := WithinScope(scope, &foldRightIncr[T, O]{
		n:   NewNode("fold_right"),
		i:   i,
		fn:  fn,
		val: v0,
	})
	Link(o, i)
	return o
}

type foldRightIncr[T, O any] struct {
	n   *Node
	i   Incr[[]T]
	fn  func(T, O) O
	val O
}

func (fli *foldRightIncr[T, O]) String() string { return fli.n.String() }

func (fli *foldRightIncr[T, O]) Node() *Node { return fli.n }

func (fli *foldRightIncr[T, O]) Value() O { return fli.val }

func (fli *foldRightIncr[T, O]) Stabilize(_ context.Context) error {
	new := fli.i.Value()
	fli.val = foldRight(new, fli.val, fli.fn)
	return nil
}

func foldRight[T, O any](input []T, v0 O, fn func(T, O) O) (o O) {
	o = v0
	for x := len(input) - 1; x >= 0; x-- {
		o = fn(input[x], o)
	}
	return
}
