package mapi

import (
	"cmp"
	"context"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// Bounds is an inclusive key range.
type Bounds[K cmp.Ordered] struct {
	Low  K
	High K
}

// Subrange restricts an incremental [pmap.Map] to a key range, where the range is
// itself incremental.
//
// This is the windowing primitive: a view scrolling over a large sorted collection
// changes its bounds far more often than the collection changes, and neither should
// cost anything proportional to the collection's size.
//
// Costs differ by what moved, and both are proportional to the thing that moved
// rather than to the map:
//
//   - the map changed: O(changes x log n), applying only the changes that fall in
//     the current range.
//   - the bounds changed: O(window), rebuilding the view by walking just the new
//     range. Subtrees outside the bounds are never visited.
func Subrange[K cmp.Ordered, V any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
	bounds incr.Incr[Bounds[K]],
	equal func(a, b V) bool,
) incr.Incr[pmap.Map[K, V]] {
	s := &subrangeIncr[K, V]{
		n:      incr.NewNode("mapi_subrange"),
		i:      input,
		bounds: bounds,
		equal:  equal,
	}
	s.parents[0] = input
	s.parents[1] = bounds
	return incr.WithinScope(scope, s)
}

var (
	_ incr.Incr[pmap.Map[string, int]] = (*subrangeIncr[string, int])(nil)
	_ incr.IStabilize                  = (*subrangeIncr[string, int])(nil)
	_ incr.IParents                    = (*subrangeIncr[string, int])(nil)
)

type subrangeIncr[K cmp.Ordered, V any] struct {
	n          *incr.Node
	i          incr.Incr[pmap.Map[K, V]]
	bounds     incr.Incr[Bounds[K]]
	equal      func(a, b V) bool
	last       pmap.Map[K, V]
	lastBounds Bounds[K]
	haveBounds bool
	value      pmap.Map[K, V]
	parents    [2]incr.INode
}

func (s *subrangeIncr[K, V]) Parents() []incr.INode { return s.parents[:] }

func (s *subrangeIncr[K, V]) Node() *incr.Node { return s.n }

func (s *subrangeIncr[K, V]) Value() pmap.Map[K, V] { return s.value }

func (s *subrangeIncr[K, V]) Stabilize(_ context.Context) error {
	current := s.i.Value()
	bounds := s.bounds.Value()

	if !s.haveBounds || bounds != s.lastBounds {
		// The window moved, so the previous view says nothing useful about the new
		// one. Rebuilding from a bounded walk costs one step per entry in the new
		// window and never looks outside it.
		out := pmap.New[K, V]()
		for key, value := range current.Range(bounds.Low, bounds.High) {
			out = out.Set(key, value)
		}
		s.value = out
		s.last = current
		s.lastBounds = bounds
		s.haveBounds = true
		return nil
	}

	// The window held still, so only the changed keys matter, and only those inside
	// it.
	out := s.value
	for change := range s.last.SymmetricDiff(current, s.equal) {
		if change.Key < bounds.Low || bounds.High < change.Key {
			continue
		}
		if change.Kind == pmap.ChangeRemoved {
			out = out.Delete(change.Key)
			continue
		}
		out = out.Set(change.Key, change.New)
	}
	s.value = out
	s.last = current
	return nil
}

func (s *subrangeIncr[K, V]) String() string { return s.n.String() }
