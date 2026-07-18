package mapi

import (
	"cmp"
	"context"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// Changes reports what changed in an incremental [pmap.Map] between stabilizations,
// as added, removed and updated entries.
//
// This is the diff itself, exposed. It costs O(changes x log n) per pass, where the
// same thing over a builtin map costs O(n): see [Added] and [Removed], which have to
// compare every key and clone the map to remember it.
func Changes[K cmp.Ordered, V any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
	equal func(a, b V) bool,
) incr.Incr[ChangeSet[K, V]] {
	c := &changesIncr[K, V]{
		n:     incr.NewNode("mapi_changes"),
		i:     input,
		equal: equal,
	}
	c.parents[0] = input
	return incr.WithinScope(scope, c)
}

// ChangeSet is what changed between two versions of a map.
//
// Updated holds the new value; the previous one is available from the diff directly via
// [pmap.Map.SymmetricDiff] if it is needed.
type ChangeSet[K cmp.Ordered, V any] struct {
	Added   pmap.Map[K, V]
	Removed pmap.Map[K, V]
	Updated pmap.Map[K, V]
}

// Len returns the total number of changes.
func (c ChangeSet[K, V]) Len() int {
	return c.Added.Len() + c.Removed.Len() + c.Updated.Len()
}

var (
	_ incr.Incr[ChangeSet[string, int]] = (*changesIncr[string, int])(nil)
	_ incr.IStabilize                   = (*changesIncr[string, int])(nil)
	_ incr.IParents                     = (*changesIncr[string, int])(nil)
)

type changesIncr[K cmp.Ordered, V any] struct {
	n       *incr.Node
	i       incr.Incr[pmap.Map[K, V]]
	equal   func(a, b V) bool
	last    pmap.Map[K, V]
	value   ChangeSet[K, V]
	parents [1]incr.INode
}

func (c *changesIncr[K, V]) Parents() []incr.INode { return c.parents[:] }

func (c *changesIncr[K, V]) Node() *incr.Node { return c.n }

func (c *changesIncr[K, V]) Value() ChangeSet[K, V] { return c.value }

func (c *changesIncr[K, V]) Stabilize(_ context.Context) error {
	current := c.i.Value()
	// a fresh set each pass: this reports what changed in *this* pass, so carrying the
	// previous one forward would accumulate rather than replace
	var next ChangeSet[K, V]
	for change := range c.last.SymmetricDiff(current, c.equal) {
		switch change.Kind {
		case pmap.ChangeAdded:
			next.Added = next.Added.Set(change.Key, change.New)
		case pmap.ChangeRemoved:
			next.Removed = next.Removed.Set(change.Key, change.Old)
		case pmap.ChangeUpdated:
			next.Updated = next.Updated.Set(change.Key, change.New)
		}
	}
	c.value = next
	c.last = current
	return nil
}

func (c *changesIncr[K, V]) String() string { return c.n.String() }
