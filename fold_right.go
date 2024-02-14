package incr

import "context"

// FoldRight folds an array from N to 0 carrying the previous value
// to the next interation, yielding a single value.
func FoldRight[T, O any](scope Scope, i Incr[[]T], v0 O, fn func(T, O) O) Incr[O] {
	return WithinScope(scope, &foldRightIncr[T, O]{
		n:   NewNode("fold_right"),
		i:   i,
		fn:  fn,
		val: v0,
	})
}

type foldRightIncr[T, O any] struct {
	n   *Node
	i   Incr[[]T]
	fn  func(T, O) O
	val O
}

func (fri *foldRightIncr[T, O]) Parents() []INode {
	return []INode{fri.i}
}

func (fri *foldRightIncr[T, O]) String() string { return fri.n.String() }

func (fri *foldRightIncr[T, O]) Node() *Node { return fri.n }

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
