package pmap

import (
	"cmp"
	"iter"
)

// ChangeKind describes how a key differs between two maps.
type ChangeKind uint8

const (
	// ChangeAdded means the key is present in the newer map only.
	ChangeAdded ChangeKind = iota
	// ChangeRemoved means the key is present in the older map only.
	ChangeRemoved
	// ChangeUpdated means the key is present in both but bound to unequal values.
	ChangeUpdated
)

func (c ChangeKind) String() string {
	switch c {
	case ChangeAdded:
		return "added"
	case ChangeRemoved:
		return "removed"
	case ChangeUpdated:
		return "updated"
	default:
		return "unknown"
	}
}

// Change is one difference between two maps.
//
// Old is meaningful for [ChangeRemoved] and [ChangeUpdated]; New for
// [ChangeAdded] and [ChangeUpdated].
type Change[K cmp.Ordered, V any] struct {
	Kind ChangeKind
	Key  K
	Old  V
	New  V
}

// SymmetricDiff iterates the keys that differ between m and other, in key order,
// treating m as the older map.
//
// Cost is proportional to the number of differences rather than to the size of
// either map, provided the two are related by updates rather than built
// separately: any subtree the two maps share is recognized by pointer and skipped
// whole, so unchanged regions cost nothing. Two independently built maps with the
// same contents share nothing and cost O(n) to compare, which is the same as
// comparing two builtin maps.
//
// equal decides whether a key present in both has changed. Pass nil to treat every
// common key as unchanged, which reports only additions and removals.
func (m Map[K, V]) SymmetricDiff(other Map[K, V], equal func(a, b V) bool) iter.Seq[Change[K, V]] {
	return func(yield func(Change[K, V]) bool) {
		diff(m.root, other.root, equal, yield)
	}
}

// diff walks two subtrees that span the same key range.
func diff[K cmp.Ordered, V any](
	older, newer *node[K, V],
	equal func(a, b V) bool,
	yield func(Change[K, V]) bool,
) bool {
	// The whole point of the structure: subtrees the two maps hold in common are
	// the same object, so an unchanged region is dismissed in one comparison.
	if older == newer {
		return true
	}
	if older == nil {
		return newer.each(func(k K, v V) bool {
			return yield(Change[K, V]{Kind: ChangeAdded, Key: k, New: v})
		})
	}
	if newer == nil {
		return older.each(func(k K, v V) bool {
			return yield(Change[K, V]{Kind: ChangeRemoved, Key: k, Old: v})
		})
	}
	// Pivot on the older subtree's root and split the newer one by that key, so
	// both sides then span matching ranges and can be compared recursively.
	// Splitting rebuilds only the spine, so subtrees hanging off it keep their
	// identity and remain skippable above.
	newLeft, newValue, newFound, newRight := split(newer, older.key)
	if !diff(older.left, newLeft, equal, yield) {
		return false
	}
	switch {
	case !newFound:
		if !yield(Change[K, V]{Kind: ChangeRemoved, Key: older.key, Old: older.value}) {
			return false
		}
	case equal != nil && !equal(older.value, newValue):
		if !yield(Change[K, V]{Kind: ChangeUpdated, Key: older.key, Old: older.value, New: newValue}) {
			return false
		}
	}
	return diff(older.right, newRight, equal, yield)
}

// split partitions a subtree around key, returning everything below it, the value
// bound to it if present, and everything above it.
func split[K cmp.Ordered, V any](n *node[K, V], key K) (left *node[K, V], value V, found bool, right *node[K, V]) {
	if n == nil {
		return nil, value, false, nil
	}
	switch {
	case key < n.key:
		l, v, ok, r := split(n.left, key)
		return l, v, ok, balance(n.key, n.value, r, n.right)
	case n.key < key:
		l, v, ok, r := split(n.right, key)
		return balance(n.key, n.value, n.left, l), v, ok, r
	default:
		return n.left, n.value, true, n.right
	}
}
