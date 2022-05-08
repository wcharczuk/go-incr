package incr

import "fmt"

// Return yields a constant incremental for a given value.
//
// Note that it does not implement `IStabilize` and is effectively
// always the same value (and treated as such).
func Return[T any](v T) Incr[T] {
	return &returnIncr[T]{
		n: NewNode(),
		v: v,
	}
}

var (
	_ Incr[string] = (*returnIncr[string])(nil)
	_ INode        = (*returnIncr[string])(nil)
	_ fmt.Stringer = (*returnIncr[string])(nil)
)

type returnIncr[T any] struct {
	n *Node
	v T
}

func (r returnIncr[T]) Node() *Node { return r.n }

func (r returnIncr[T]) Value() T { return r.v }

func (r returnIncr[T]) String() string { return FormatNode(r.n, "return") }
