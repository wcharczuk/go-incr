package pmap

import "cmp"

// Reducer folds a [Map] to a single value, reusing the work it did for the parts of
// the map that have not changed since the last fold.
//
// This is how to maintain an aggregate that has no inverse. [Map.SymmetricDiff] lets
// a caller adjust a running total from the entries that changed, but that only works
// when a contribution can be withdrawn: given a maximum, removing the current maximum
// says nothing about the next one. Folding over the tree instead means a changed key
// only invalidates the subtrees on its path to the root, so a fold costs
// O(changes x log n) while still visiting whatever it needs.
//
// A Reducer is stateful and not safe for concurrent use. It is tied to the sequence
// of maps it is called with, and gains nothing when called on unrelated maps.
//
// combine must be associative. It need not be commutative: subtrees are combined in
// key order.
type Reducer[K cmp.Ordered, V, R any] struct {
	project func(K, V) R
	combine func(R, R) R
	// memo holds the fold of each subtree, keyed by the subtree itself. Because an
	// update rebuilds only the path to the changed key, every other subtree is the
	// same object as before and is found here.
	memo map[*node[K, V]]R
	// reachable is reused when pruning, to avoid allocating a set per prune.
	reachable map[*node[K, V]]struct{}
}

// NewReducer returns a Reducer that projects each entry with project and combines
// results with combine.
func NewReducer[K cmp.Ordered, V, R any](project func(K, V) R, combine func(R, R) R) *Reducer[K, V, R] {
	return &Reducer[K, V, R]{
		project: project,
		combine: combine,
		memo:    make(map[*node[K, V]]R),
	}
}

// Reduce folds a map, returning false if the map is empty.
func (r *Reducer[K, V, R]) Reduce(m Map[K, V]) (R, bool) {
	// Subtrees the map has moved past stay in the memo until they are pruned, since
	// nothing reports when a version is discarded. Pruning is proportional to the
	// map, so it is amortized against the growth that triggered it. Live folds already
	// occupy roughly one entry per node, so the threshold below waits for about three
	// times that many stale entries to accumulate before paying for a sweep.
	if len(r.memo) > 4*m.Len()+64 {
		r.prune(m.root)
	}
	return r.reduce(m.root)
}

func (r *Reducer[K, V, R]) reduce(n *node[K, V]) (R, bool) {
	var zero R
	if n == nil {
		return zero, false
	}
	// The whole point: an unchanged subtree is the same object as last time, so its
	// fold is already known and the walk stops here.
	if cached, ok := r.memo[n]; ok {
		return cached, true
	}
	// combine in key order, so a non-commutative operation is well defined
	acc := r.project(n.key, n.value)
	if left, ok := r.reduce(n.left); ok {
		acc = r.combine(left, acc)
	}
	if right, ok := r.reduce(n.right); ok {
		acc = r.combine(acc, right)
	}
	r.memo[n] = acc
	return acc, true
}

// prune drops memo entries for subtrees no longer reachable from the live map.
func (r *Reducer[K, V, R]) prune(root *node[K, V]) {
	if r.reachable == nil {
		r.reachable = make(map[*node[K, V]]struct{}, len(r.memo))
	}
	clear(r.reachable)
	var walk func(*node[K, V])
	walk = func(n *node[K, V]) {
		if n == nil {
			return
		}
		r.reachable[n] = struct{}{}
		walk(n.left)
		walk(n.right)
	}
	walk(root)
	for n := range r.memo {
		if _, ok := r.reachable[n]; !ok {
			delete(r.memo, n)
		}
	}
}

// MemoLen returns the number of subtree folds currently retained, which is useful
// for confirming the memo stays bounded as a map is updated repeatedly.
func (r *Reducer[K, V, R]) MemoLen() int { return len(r.memo) }
