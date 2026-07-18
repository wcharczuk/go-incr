package mapi

import (
	"cmp"
	"context"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// UnorderedFold folds an incremental [pmap.Map] into a single value, adjusting the
// result from the entries that changed rather than re-folding.
//
// add applies an entry's contribution to the accumulator and remove withdraws it,
// so a changed key is handled by removing its old contribution and adding its new
// one. A pass therefore costs O(changes x log n) rather than O(n), which is what
// makes an aggregate over a large map practical.
//
// This is the keyed counterpart of [incr.UnorderedArrayFold], and carries the same
// requirement: the operation needs an inverse. For a sum over values,
//
//	add:    func(acc int, _ K, v int) int { return acc + v }
//	remove: func(acc int, _ K, v int) int { return acc - v }
//
// An aggregate with no inverse, such as a maximum, cannot be maintained this way.
//
// The fold order is unspecified, hence "unordered": with a correct add and remove
// the accumulator does not depend on it.
func UnorderedFold[K cmp.Ordered, V, B any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
	initial B,
	equal func(a, b V) bool,
	add func(acc B, key K, value V) B,
	remove func(acc B, key K, value V) B,
) incr.Incr[B] {
	f := &unorderedFoldIncr[K, V, B]{
		n:      incr.NewNode("mapi_unordered_fold"),
		i:      input,
		equal:  equal,
		add:    add,
		remove: remove,
		value:  initial,
	}
	f.parents[0] = input
	return incr.WithinScope(scope, f)
}

var (
	_ incr.Incr[int]  = (*unorderedFoldIncr[string, int, int])(nil)
	_ incr.IStabilize = (*unorderedFoldIncr[string, int, int])(nil)
	_ incr.IParents   = (*unorderedFoldIncr[string, int, int])(nil)
)

type unorderedFoldIncr[K cmp.Ordered, V, B any] struct {
	n       *incr.Node
	i       incr.Incr[pmap.Map[K, V]]
	equal   func(a, b V) bool
	add     func(B, K, V) B
	remove  func(B, K, V) B
	last    pmap.Map[K, V]
	value   B
	parents [1]incr.INode
}

func (f *unorderedFoldIncr[K, V, B]) Parents() []incr.INode { return f.parents[:] }

func (f *unorderedFoldIncr[K, V, B]) Node() *incr.Node { return f.n }

func (f *unorderedFoldIncr[K, V, B]) Value() B { return f.value }

func (f *unorderedFoldIncr[K, V, B]) Stabilize(_ context.Context) error {
	current := f.i.Value()
	acc := f.value
	for change := range f.last.SymmetricDiff(current, f.equal) {
		switch change.Kind {
		case pmap.ChangeAdded:
			acc = f.add(acc, change.Key, change.New)
		case pmap.ChangeRemoved:
			acc = f.remove(acc, change.Key, change.Old)
		case pmap.ChangeUpdated:
			acc = f.remove(acc, change.Key, change.Old)
			acc = f.add(acc, change.Key, change.New)
		}
	}
	f.value = acc
	f.last = current
	return nil
}

func (f *unorderedFoldIncr[K, V, B]) String() string { return f.n.String() }
