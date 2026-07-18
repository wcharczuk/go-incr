package mapi

import (
	"cmp"
	"context"
	"slices"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// MergeElement is what a key looks like across two maps being merged: present on
// the left, the right, or both.
type MergeElement[A, B any] struct {
	Left     A
	HasLeft  bool
	Right    B
	HasRight bool
}

// Merge combines two incremental [pmap.Map] values into one, recomputing only the
// keys that changed in either input.
//
// fn sees each key together with whichever sides hold it, and returns the merged
// value and whether to include the key. A key absent from both inputs is not
// visited.
//
// Each pass diffs both inputs against what they were last time and recomputes the
// union of the changed keys, so cost is O(changes x log n) in the number of changes
// across both inputs rather than O(n) in their size.
func Merge[K cmp.Ordered, A, B, C any](
	scope incr.Scope,
	left incr.Incr[pmap.Map[K, A]],
	right incr.Incr[pmap.Map[K, B]],
	equalLeft func(a, b A) bool,
	equalRight func(a, b B) bool,
	fn func(K, MergeElement[A, B]) (C, bool),
) incr.Incr[pmap.Map[K, C]] {
	m := &mergeIncr[K, A, B, C]{
		n:          incr.NewNode("mapi_merge"),
		left:       left,
		right:      right,
		equalLeft:  equalLeft,
		equalRight: equalRight,
		fn:         fn,
	}
	m.parents[0] = left
	m.parents[1] = right
	return incr.WithinScope(scope, m)
}

var (
	_ incr.Incr[pmap.Map[string, int]] = (*mergeIncr[string, int, int, int])(nil)
	_ incr.IStabilize                  = (*mergeIncr[string, int, int, int])(nil)
	_ incr.IParents                    = (*mergeIncr[string, int, int, int])(nil)
)

type mergeIncr[K cmp.Ordered, A, B, C any] struct {
	n          *incr.Node
	left       incr.Incr[pmap.Map[K, A]]
	right      incr.Incr[pmap.Map[K, B]]
	equalLeft  func(a, b A) bool
	equalRight func(a, b B) bool
	fn         func(K, MergeElement[A, B]) (C, bool)
	lastLeft   pmap.Map[K, A]
	lastRight  pmap.Map[K, B]
	value      pmap.Map[K, C]
	parents    [2]incr.INode
	// touched is reused between passes to collect the keys needing recomputation,
	// so a pass allocates nothing when the change set is small.
	touched []K
}

func (m *mergeIncr[K, A, B, C]) Parents() []incr.INode { return m.parents[:] }

func (m *mergeIncr[K, A, B, C]) Node() *incr.Node { return m.n }

func (m *mergeIncr[K, A, B, C]) Value() pmap.Map[K, C] { return m.value }

func (m *mergeIncr[K, A, B, C]) Stabilize(_ context.Context) error {
	currentLeft, currentRight := m.left.Value(), m.right.Value()

	// A key may have changed on either side, or on both; collecting first keeps it
	// from being recomputed twice.
	m.touched = m.touched[:0]
	for change := range m.lastLeft.SymmetricDiff(currentLeft, m.equalLeft) {
		m.touched = append(m.touched, change.Key)
	}
	for change := range m.lastRight.SymmetricDiff(currentRight, m.equalRight) {
		m.touched = append(m.touched, change.Key)
	}

	out := m.value
	// both diffs yield keys in order, so a duplicate can only be the key seen
	// immediately before it once the two runs are merged; rather than rely on that,
	// track the previous key and skip repeats after sorting.
	slices.Sort(m.touched)
	var previous K
	for index, key := range m.touched {
		if index > 0 && key == previous {
			continue
		}
		previous = key

		var element MergeElement[A, B]
		element.Left, element.HasLeft = currentLeft.Get(key)
		element.Right, element.HasRight = currentRight.Get(key)
		if !element.HasLeft && !element.HasRight {
			out = out.Delete(key)
			continue
		}
		merged, include := m.fn(key, element)
		if include {
			out = out.Set(key, merged)
		} else {
			out = out.Delete(key)
		}
	}

	m.value = out
	m.lastLeft = currentLeft
	m.lastRight = currentRight
	return nil
}

func (m *mergeIncr[K, A, B, C]) String() string { return m.n.String() }
