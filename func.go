package incr

import (
	"context"
	"fmt"
)

// Func wraps a given function as an incremental.
//
// The result of the function after the first stabilization will
// be re-used between stabilizations unless you mark the node stale
// with the `SetStale` function on the `Graph` type.
//
// Because there is no tracking of input changes, this node
// type is generally discouraged in favor of `Map` or `Bind`
// incrementals but is included for "expert" use cases, typically
// as an input to other nodes.
func Func[T any](scope Scope, fn func(context.Context) (T, error)) Incr[T] {
	return WithinScope(scope, &funcIncr[T]{
		n:  NewNode("func"),
		fn: fn,
	})
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

func (f *funcIncr[T]) Parents() []INode { return nil }

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
	return f.n.String()
}
