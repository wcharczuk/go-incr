// Package pmap provides an immutable ordered map with structural sharing.
//
// Updating a [Map] returns a new map that shares every subtree it did not have to
// rebuild, which is what makes [Map.SymmetricDiff] cheap: two maps related by a
// few updates differ only along the paths to the changed keys, and the diff skips
// any subtree the two maps hold in common by comparing pointers.
//
// That is the property incremental map operations need. Comparing two of Go's
// builtin maps costs O(n) however few keys changed, because a hash table shares no
// structure with the map it was copied from and offers no way to ask what differs.
// A computation keyed on a map therefore cannot update in proportion to the change
// unless the map itself can be diffed cheaply.
//
// The zero Map is a valid empty map. Values are compared by a caller-supplied
// equality function where it matters, since Go has no structural equality for
// arbitrary types.
package pmap

import (
	"cmp"
	"iter"
)

// Map is an immutable ordered map from K to V.
//
// A Map value is a handle onto a shared tree; copying one is cheap and the copies
// are independent, since no operation mutates a tree that is already reachable.
type Map[K cmp.Ordered, V any] struct {
	root *node[K, V]
}

// node is a tree node. Once published a node is never mutated, which is what makes
// sharing safe: a subtree can appear in any number of maps at once.
type node[K cmp.Ordered, V any] struct {
	key         K
	value       V
	left, right *node[K, V]
	// height and size are maintained for balancing and for O(1) Len.
	height int
	size   int
}

// New returns an empty map. The zero value is equally valid.
func New[K cmp.Ordered, V any]() Map[K, V] {
	return Map[K, V]{}
}

// Len returns the number of entries, in constant time.
func (m Map[K, V]) Len() int { return m.root.treeSize() }

// Get returns the value stored under key.
func (m Map[K, V]) Get(key K) (value V, ok bool) {
	cursor := m.root
	for cursor != nil {
		switch {
		case key < cursor.key:
			cursor = cursor.left
		case cursor.key < key:
			cursor = cursor.right
		default:
			return cursor.value, true
		}
	}
	return
}

// Has reports whether key is present.
func (m Map[K, V]) Has(key K) bool {
	_, ok := m.Get(key)
	return ok
}

// Set returns a map with key bound to value, leaving the receiver unchanged.
//
// Only the nodes on the path to key are rebuilt; every other subtree is shared
// with the receiver.
func (m Map[K, V]) Set(key K, value V) Map[K, V] {
	return Map[K, V]{root: insert(m.root, key, value)}
}

// Delete returns a map without key, leaving the receiver unchanged.
func (m Map[K, V]) Delete(key K) Map[K, V] {
	return Map[K, V]{root: remove(m.root, key)}
}

// All iterates entries in key order.
func (m Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.root.each(yield)
	}
}

// Keys iterates keys in order.
func (m Map[K, V]) Keys() iter.Seq[K] {
	return func(yield func(K) bool) {
		m.root.each(func(k K, _ V) bool { return yield(k) })
	}
}

//
// tree internals
//

func (n *node[K, V]) treeSize() int {
	if n == nil {
		return 0
	}
	return n.size
}

func (n *node[K, V]) treeHeight() int {
	if n == nil {
		return 0
	}
	return n.height
}

func (n *node[K, V]) each(yield func(K, V) bool) bool {
	if n == nil {
		return true
	}
	if !n.left.each(yield) {
		return false
	}
	if !yield(n.key, n.value) {
		return false
	}
	return n.right.each(yield)
}

// join builds a node from a key, value and two balanced subtrees whose heights
// differ by at most one, computing the derived fields.
func join[K cmp.Ordered, V any](key K, value V, left, right *node[K, V]) *node[K, V] {
	height := left.treeHeight()
	if rh := right.treeHeight(); rh > height {
		height = rh
	}
	return &node[K, V]{
		key:    key,
		value:  value,
		left:   left,
		right:  right,
		height: height + 1,
		size:   left.treeSize() + right.treeSize() + 1,
	}
}

// balance builds a node from subtrees whose heights may differ by two, rotating as
// needed. Rotations allocate rather than mutate, so trees already published are
// untouched.
func balance[K cmp.Ordered, V any](key K, value V, left, right *node[K, V]) *node[K, V] {
	lh, rh := left.treeHeight(), right.treeHeight()
	switch {
	case lh > rh+1:
		if left.left.treeHeight() >= left.right.treeHeight() {
			return join(left.key, left.value, left.left, join(key, value, left.right, right))
		}
		lr := left.right
		return join(lr.key, lr.value,
			join(left.key, left.value, left.left, lr.left),
			join(key, value, lr.right, right))
	case rh > lh+1:
		if right.right.treeHeight() >= right.left.treeHeight() {
			return join(right.key, right.value, join(key, value, left, right.left), right.right)
		}
		rl := right.left
		return join(rl.key, rl.value,
			join(key, value, left, rl.left),
			join(right.key, right.value, rl.right, right.right))
	default:
		return join(key, value, left, right)
	}
}

func insert[K cmp.Ordered, V any](n *node[K, V], key K, value V) *node[K, V] {
	if n == nil {
		return join(key, value, nil, nil)
	}
	switch {
	case key < n.key:
		return balance(n.key, n.value, insert(n.left, key, value), n.right)
	case n.key < key:
		return balance(n.key, n.value, n.left, insert(n.right, key, value))
	default:
		// rebinding an existing key rebuilds only this node, so both subtrees stay
		// shared with the receiver.
		return join(key, value, n.left, n.right)
	}
}

func remove[K cmp.Ordered, V any](n *node[K, V], key K) *node[K, V] {
	if n == nil {
		return nil
	}
	switch {
	case key < n.key:
		return balance(n.key, n.value, remove(n.left, key), n.right)
	case n.key < key:
		return balance(n.key, n.value, n.left, remove(n.right, key))
	default:
		return glue(n.left, n.right)
	}
}

// glue joins two subtrees that sat under a removed node, by promoting the
// in-order neighbor from whichever side is taller.
func glue[K cmp.Ordered, V any](left, right *node[K, V]) *node[K, V] {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	if left.treeHeight() > right.treeHeight() {
		key, value, rest := removeMax(left)
		return balance(key, value, rest, right)
	}
	key, value, rest := removeMin(right)
	return balance(key, value, left, rest)
}

func removeMin[K cmp.Ordered, V any](n *node[K, V]) (K, V, *node[K, V]) {
	if n.left == nil {
		return n.key, n.value, n.right
	}
	key, value, rest := removeMin(n.left)
	return key, value, balance(n.key, n.value, rest, n.right)
}

func removeMax[K cmp.Ordered, V any](n *node[K, V]) (K, V, *node[K, V]) {
	if n.right == nil {
		return n.key, n.value, n.left
	}
	key, value, rest := removeMax(n.right)
	return key, value, balance(n.key, n.value, n.left, rest)
}

// Min returns the smallest key and its value.
func (m Map[K, V]) Min() (key K, value V, ok bool) {
	cursor := m.root
	if cursor == nil {
		return
	}
	for cursor.left != nil {
		cursor = cursor.left
	}
	return cursor.key, cursor.value, true
}

// Max returns the largest key and its value.
func (m Map[K, V]) Max() (key K, value V, ok bool) {
	cursor := m.root
	if cursor == nil {
		return
	}
	for cursor.right != nil {
		cursor = cursor.right
	}
	return cursor.key, cursor.value, true
}

// Range iterates the entries with keys in [low, high], in order.
//
// Subtrees wholly outside the bounds are skipped, so this costs O(log n) to reach
// the range plus one step per entry in it, rather than a walk of the whole map.
func (m Map[K, V]) Range(low, high K) iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		m.root.eachRange(low, high, yield)
	}
}

func (n *node[K, V]) eachRange(low, high K, yield func(K, V) bool) bool {
	if n == nil {
		return true
	}
	// only descend left if something there can be at or above low
	if low < n.key {
		if !n.left.eachRange(low, high, yield) {
			return false
		}
	}
	if !(n.key < low) && !(high < n.key) {
		if !yield(n.key, n.value) {
			return false
		}
	}
	// and right only if something there can be at or below high
	if n.key < high {
		return n.right.eachRange(low, high, yield)
	}
	return true
}
