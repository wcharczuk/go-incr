package mapi

import (
	"cmp"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// Cardinality returns the number of entries in an incremental [pmap.Map].
//
// Maintained from the diff, so it costs O(changes x log n) per pass rather than
// reading the map.
func Cardinality[K cmp.Ordered, V any](scope incr.Scope, input incr.Incr[pmap.Map[K, V]]) incr.Incr[int] {
	// no equality function: a rebound value does not change the count, so there is
	// nothing to recompute when one changes.
	return UnorderedFold(scope, input, 0, nil,
		func(acc int, _ K, _ V) int { return acc + 1 },
		func(acc int, _ K, _ V) int { return acc - 1 })
}

// Counti returns the number of entries satisfying a predicate.
//
// equal decides when a value has changed enough to re-test it.
func Counti[K cmp.Ordered, V any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
	equal func(a, b V) bool,
	predicate func(K, V) bool,
) incr.Incr[int] {
	count := func(acc int, key K, value V, delta int) int {
		if predicate(key, value) {
			return acc + delta
		}
		return acc
	}
	return UnorderedFold(scope, input, 0, equal,
		func(acc int, key K, value V) int { return count(acc, key, value, 1) },
		func(acc int, key K, value V) int { return count(acc, key, value, -1) })
}

// Sum returns the sum of an incremental map's values.
//
// Addition has an inverse, so this is maintained in constant time per changed key.
//
// An aggregate with no inverse cannot be built this way. A maximum over values is the
// usual example: withdrawing the current maximum leaves no way to know the next one
// without looking at the rest. Maintaining one needs a reduction shaped like the tree
// itself, so that removing a value recomputes only its path to the root, and that is
// not provided here yet -- an implementation that rescans on removal would be O(n)
// in the worst case, which is the kind of cliff these operators exist to avoid.
// [pmap.Map.Min] and [pmap.Map.Max] give the extremes by *key* in O(log n).
func Sum[K cmp.Ordered, V interface {
	~int | ~int64 | ~float64
}](scope incr.Scope, input incr.Incr[pmap.Map[K, V]], equal func(a, b V) bool) incr.Incr[V] {
	var zero V
	return UnorderedFold(scope, input, zero, equal,
		func(acc V, _ K, value V) V { return acc + value },
		func(acc V, _ K, value V) V { return acc - value })
}

// Keys returns an incremental map's keys in order.
//
// The result is rebuilt from the map on any change, which costs O(n); a sorted key
// list cannot be maintained incrementally as a plain slice. Prefer iterating the map
// itself where the ordered slice is not actually needed.
func Keys[K cmp.Ordered, V any](scope incr.Scope, input incr.Incr[pmap.Map[K, V]]) incr.Incr[[]K] {
	return incr.Map(scope, input, func(m pmap.Map[K, V]) []K {
		out := make([]K, 0, m.Len())
		for key := range m.Keys() {
			out = append(out, key)
		}
		return out
	})
}
