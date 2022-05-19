package incr

import (
	"context"
	"fmt"
)

// Func wraps a given function as an incremental.
//
// You can mark this for recomputation with the `SetStale` helper.
func Func[T any](fn func(context.Context) (T, error)) Incr[T] {
	return &funcIncr[T]{
		n:  NewNode(),
		fn: fn,
	}
}

var (
	_ Incr[string] = (*funcIncr[string])(nil)
	_ INode        = (*funcIncr[string])(nil)
	_ IStabilize   = (*funcIncr[string])(nil)
	_ fmt.Stringer = (*funcIncr[string])(nil)
)

type funcIncr[T any] struct {
	n   *Node
	fn  func(context.Context) (T, error)
	val T
}

func (f *funcIncr[T]) Node() *Node { return f.n }
func (f *funcIncr[T]) Value() T    { return f.val }
func (f *funcIncr[T]) Stabilize(ctx context.Context) error {
	val, err := f.fn(ctx)
	if err != nil {
		return err
	}
	f.val = val
	return nil
}

func (f *funcIncr[T]) String() string {
	return Label(f.n, "func")
}
