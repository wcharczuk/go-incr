package mapi

import (
	"cmp"
	"context"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// Optional carries a value that may be absent, which is how an aggregate over a
// possibly-empty map reports having nothing to report.
type Optional[V any] struct {
	Value   V
	Present bool
}

// Reduce folds an incremental [pmap.Map] to a single value with an operation that
// need not have an inverse.
//
// [UnorderedFold] is cheaper when the operation has one, since it adjusts a running
// total from the entries that changed. When it does not -- a maximum, a minimum, a
// concatenation -- there is no way to withdraw a contribution, and this is the
// alternative: the fold is computed over the map's own tree, so a changed key
// invalidates only the subtrees on its path to the root and a pass costs
// O(changes x log n).
//
// project turns an entry into the value being combined; combine must be associative.
// It need not be commutative, because subtrees are combined in key order.
//
// empty is the result for an empty map, and combine is never called with it.
func Reduce[K cmp.Ordered, V, R any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
	empty R,
	project func(K, V) R,
	combine func(R, R) R,
) incr.Incr[R] {
	r := &reduceIncr[K, V, R]{
		n:       incr.NewNode("mapi_reduce"),
		i:       input,
		empty:   empty,
		reducer: pmap.NewReducer(project, combine),
	}
	r.value = empty
	r.parents[0] = input
	return incr.WithinScope(scope, r)
}

var (
	_ incr.Incr[int]  = (*reduceIncr[string, int, int])(nil)
	_ incr.IStabilize = (*reduceIncr[string, int, int])(nil)
	_ incr.IParents   = (*reduceIncr[string, int, int])(nil)
)

type reduceIncr[K cmp.Ordered, V, R any] struct {
	n       *incr.Node
	i       incr.Incr[pmap.Map[K, V]]
	empty   R
	reducer *pmap.Reducer[K, V, R]
	value   R
	parents [1]incr.INode
}

func (r *reduceIncr[K, V, R]) Parents() []incr.INode { return r.parents[:] }

func (r *reduceIncr[K, V, R]) Node() *incr.Node { return r.n }

func (r *reduceIncr[K, V, R]) Value() R { return r.value }

func (r *reduceIncr[K, V, R]) Stabilize(_ context.Context) error {
	if value, ok := r.reducer.Reduce(r.i.Value()); ok {
		r.value = value
		return nil
	}
	r.value = r.empty
	return nil
}

func (r *reduceIncr[K, V, R]) String() string { return r.n.String() }

// MaxValue returns the largest value in an incremental map, absent when the map is
// empty.
//
// A maximum has no inverse, so this is built on [Reduce] rather than [UnorderedFold].
func MaxValue[K cmp.Ordered, V cmp.Ordered](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
) incr.Incr[Optional[V]] {
	return Reduce(scope, input, Optional[V]{},
		func(_ K, value V) Optional[V] { return Optional[V]{Value: value, Present: true} },
		func(a, b Optional[V]) Optional[V] {
			if b.Value > a.Value {
				return b
			}
			return a
		})
}

// MinValue returns the smallest value in an incremental map, absent when the map is
// empty. See [MaxValue].
func MinValue[K cmp.Ordered, V cmp.Ordered](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
) incr.Incr[Optional[V]] {
	return Reduce(scope, input, Optional[V]{},
		func(_ K, value V) Optional[V] { return Optional[V]{Value: value, Present: true} },
		func(a, b Optional[V]) Optional[V] {
			if b.Value < a.Value {
				return b
			}
			return a
		})
}
