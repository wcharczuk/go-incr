package mapi

import (
	"cmp"
	"context"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// MapValues applies fn to every value of an incremental [pmap.Map], recomputing
// only the entries that changed.
//
// This is the operation the persistent representation exists for. Each pass diffs
// the input against what it saw last time, and because the two maps share every
// subtree the input did not rebuild, the diff and therefore this node cost O(changes
// x log n) rather than O(n). Doing the same over a builtin map means comparing all
// the keys, which is O(n) however few of them moved.
//
// fn is called for a key when it is added or its value changes, and not at all for
// keys that stayed put. equal decides what "changed" means for an input value; pass
// nil to recompute a key only when it is added, never when rebound.
func MapValues[K cmp.Ordered, V, W any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
	equal func(a, b V) bool,
	fn func(K, V) W,
) incr.Incr[pmap.Map[K, W]] {
	m := &mapValuesIncr[K, V, W]{
		n:     incr.NewNode("mapi_map_values"),
		i:     input,
		equal: equal,
		fn:    fn,
	}
	m.parents[0] = input
	return incr.WithinScope(scope, m)
}

var (
	_ incr.Incr[pmap.Map[string, int]] = (*mapValuesIncr[string, int, int])(nil)
	_ incr.IStabilize                  = (*mapValuesIncr[string, int, int])(nil)
	_ incr.IParents                    = (*mapValuesIncr[string, int, int])(nil)
)

type mapValuesIncr[K cmp.Ordered, V, W any] struct {
	n     *incr.Node
	i     incr.Incr[pmap.Map[K, V]]
	equal func(a, b V) bool
	fn    func(K, V) W
	// last is the input as of the previous pass, retained so the next pass can be
	// diffed against it. Holding it costs nothing beyond the nodes it already
	// shares with the current input.
	last pmap.Map[K, V]
	// value is carried forward between passes and edited in place of the changed
	// keys, rather than rebuilt.
	value pmap.Map[K, W]
	// parents is an array so constructing the node does not allocate an input list.
	parents [1]incr.INode
}

func (m *mapValuesIncr[K, V, W]) Parents() []incr.INode { return m.parents[:] }

func (m *mapValuesIncr[K, V, W]) Node() *incr.Node { return m.n }

func (m *mapValuesIncr[K, V, W]) Value() pmap.Map[K, W] { return m.value }

func (m *mapValuesIncr[K, V, W]) Stabilize(_ context.Context) error {
	current := m.i.Value()
	out := m.value
	for change := range m.last.SymmetricDiff(current, m.equal) {
		switch change.Kind {
		case pmap.ChangeRemoved:
			out = out.Delete(change.Key)
		default:
			out = out.Set(change.Key, m.fn(change.Key, change.New))
		}
	}
	m.value = out
	m.last = current
	return nil
}

func (m *mapValuesIncr[K, V, W]) String() string { return m.n.String() }
