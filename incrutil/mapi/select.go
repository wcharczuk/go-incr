package mapi

import (
	"cmp"
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// Selector hands out an incremental per key of a shared map, where a change to one key
// only recomputes the consumers of that key.
//
// This is the shape where a collection has many independent watchers: a table with a row
// per key and a view per row, a set of subscriptions each following one instrument, a
// cache with a consumer per entry. Done directly -- every consumer taking the whole map
// and picking its key out -- a single key changing recomputes all of them, so the cost of
// a change scales with how many watchers there are rather than with what changed.
//
// A Selector inverts that. It watches the map once, works out which keys moved, and wakes
// only the per-key nodes for those keys. One key changing costs the consumers of that key
// and nothing else, however many other keys are being watched.
type Selector[K cmp.Ordered, V any] struct {
	scope    incr.Scope
	input    incr.Incr[pmap.Map[K, V]]
	equal    func(a, b V) bool
	fanout   *selectorIncr[K, V]
	selected map[K]*selectedIncr[K, V]
}

// NewSelector returns a Selector over an incremental map.
//
// equal decides whether a key's value has changed; pass nil to treat any key present in
// both the old and new map as unchanged.
func NewSelector[K cmp.Ordered, V any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
	equal func(a, b V) bool,
) *Selector[K, V] {
	s := &Selector[K, V]{
		scope:    scope,
		input:    input,
		equal:    equal,
		selected: make(map[K]*selectedIncr[K, V]),
	}
	s.fanout = &selectorIncr[K, V]{owner: s, n: incr.NewNode("mapi_selector")}
	s.fanout.parents[0] = input
	incr.WithinScope(scope, s.fanout)
	return s
}

// Select returns the incremental for a key, creating it on first use.
//
// The result holds the key's current value, and the zero value while the key is absent.
// Asking for the same key twice returns the same node, so consumers of a key share one.
func (s *Selector[K, V]) Select(key K) incr.Incr[V] {
	if existing, ok := s.selected[key]; ok {
		return existing
	}
	node := &selectedIncr[K, V]{owner: s, key: key, n: incr.NewNode("mapi_selected")}
	// The per-key node takes the fan-out node as its input, which puts it above the
	// fan-out in height and so after it in a pass. That alone would make every per-key
	// node stale whenever the map changed, which is the cost this type exists to avoid;
	// the dirty flag below is what narrows it.
	node.parents[0] = s.fanout
	incr.WithinScope(s.scope, node)
	s.selected[key] = node
	return node
}

var (
	_ incr.Incr[int]  = (*selectorIncr[string, int])(nil)
	_ incr.IStabilize = (*selectorIncr[string, int])(nil)
	_ incr.IParents   = (*selectorIncr[string, int])(nil)
)

// selectorIncr watches the map and marks the per-key nodes that need to recompute.
type selectorIncr[K cmp.Ordered, V any] struct {
	owner   *Selector[K, V]
	n       *incr.Node
	last    pmap.Map[K, V]
	current pmap.Map[K, V]
	parents [1]incr.INode
}

func (s *selectorIncr[K, V]) Parents() []incr.INode { return s.parents[:] }

func (s *selectorIncr[K, V]) Node() *incr.Node { return s.n }

// Value reports the number of keys, so the node has something meaningful to hand back;
// consumers read values through the per-key nodes rather than from here.
func (s *selectorIncr[K, V]) Value() int { return s.current.Len() }

func (s *selectorIncr[K, V]) Stabilize(_ context.Context) error {
	s.current = s.owner.input.Value()
	// Only the keys that actually moved are marked, so the fan-out is proportional to the
	// change rather than to the number of watchers.
	for change := range s.last.SymmetricDiff(s.current, s.owner.equal) {
		if node, ok := s.owner.selected[change.Key]; ok {
			node.dirty = true
		}
	}
	s.last = s.current
	return nil
}

func (s *selectorIncr[K, V]) String() string { return s.n.String() }

var (
	_ incr.Incr[int]  = (*selectedIncr[string, int])(nil)
	_ incr.IStabilize = (*selectedIncr[string, int])(nil)
	_ incr.IParents   = (*selectedIncr[string, int])(nil)
	_ incr.IStale     = (*selectedIncr[string, int])(nil)
	_ fmt.Stringer    = (*selectedIncr[string, int])(nil)
)

// selectedIncr is one key's view of the map.
type selectedIncr[K cmp.Ordered, V any] struct {
	owner   *Selector[K, V]
	n       *incr.Node
	key     K
	value   V
	dirty   bool
	seeded  bool
	parents [1]incr.INode
}

func (s *selectedIncr[K, V]) Parents() []incr.INode { return s.parents[:] }

func (s *selectedIncr[K, V]) Node() *incr.Node { return s.n }

func (s *selectedIncr[K, V]) Value() V { return s.value }

// Stale reports whether this key in particular needs recomputing.
//
// This is what narrows the fan-out. Without it the node would be stale simply because its
// input -- the shared fan-out node -- recomputed, which happens whenever any key changes.
func (s *selectedIncr[K, V]) Stale() bool { return s.dirty || !s.seeded }

func (s *selectedIncr[K, V]) Stabilize(_ context.Context) error {
	var zero V
	value, ok := s.owner.fanout.current.Get(s.key)
	if !ok {
		value = zero
	}
	s.value = value
	s.dirty = false
	s.seeded = true
	return nil
}

func (s *selectedIncr[K, V]) String() string { return s.n.String() }
