package mapi

import (
	"cmp"
	"context"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// FilterMapValues applies fn to every value of an incremental [pmap.Map], keeping
// only the entries fn accepts, and recomputes only the entries that changed.
//
// This is [MapValues] with the option to drop a key: fn returns the mapped value
// and whether to include it. A key whose result flips from included to excluded is
// removed from the output, so the output tracks the predicate as well as the values.
//
// Cost is O(changes x log n) per pass rather than O(n); see [MapValues] for why.
func FilterMapValues[K cmp.Ordered, V, W any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
	equal func(a, b V) bool,
	fn func(K, V) (W, bool),
) incr.Incr[pmap.Map[K, W]] {
	m := &filterMapValuesIncr[K, V, W]{
		n:     incr.NewNode("mapi_filter_map_values"),
		i:     input,
		equal: equal,
		fn:    fn,
	}
	m.parents[0] = input
	return incr.WithinScope(scope, m)
}

var (
	_ incr.Incr[pmap.Map[string, int]] = (*filterMapValuesIncr[string, int, int])(nil)
	_ incr.IStabilize                  = (*filterMapValuesIncr[string, int, int])(nil)
	_ incr.IParents                    = (*filterMapValuesIncr[string, int, int])(nil)
)

type filterMapValuesIncr[K cmp.Ordered, V, W any] struct {
	n       *incr.Node
	i       incr.Incr[pmap.Map[K, V]]
	equal   func(a, b V) bool
	fn      func(K, V) (W, bool)
	last    pmap.Map[K, V]
	value   pmap.Map[K, W]
	parents [1]incr.INode
}

func (m *filterMapValuesIncr[K, V, W]) Parents() []incr.INode { return m.parents[:] }

func (m *filterMapValuesIncr[K, V, W]) Node() *incr.Node { return m.n }

func (m *filterMapValuesIncr[K, V, W]) Value() pmap.Map[K, W] { return m.value }

func (m *filterMapValuesIncr[K, V, W]) Stabilize(_ context.Context) error {
	current := m.i.Value()
	out := m.value
	for change := range m.last.SymmetricDiff(current, m.equal) {
		if change.Kind == pmap.ChangeRemoved {
			out = out.Delete(change.Key)
			continue
		}
		mapped, include := m.fn(change.Key, change.New)
		if include {
			out = out.Set(change.Key, mapped)
		} else {
			// a key that was included before and is not now has to leave the output
			out = out.Delete(change.Key)
		}
	}
	m.value = out
	m.last = current
	return nil
}

func (m *filterMapValuesIncr[K, V, W]) String() string { return m.n.String() }
