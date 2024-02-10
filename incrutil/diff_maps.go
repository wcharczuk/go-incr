package incrutil

import (
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
)

// DiffMapByKeys returns two incrementals, one for keys added, and one
// for keys removed, and each stabilization pass returns just the subset
// of the map that changed since the last pass according to the keys.
func DiffMapByKeys[K comparable, V any](scope *incr.BindScope, i incr.Incr[map[K]V]) (add incr.Incr[map[K]V], rem incr.Incr[map[K]V]) {
	add = &diffMapByKeysAddedIncr[K, V]{
		n: incr.NewNode("diff_maps_by_keys_added"),
		i: i,
	}
	incr.Link(add, i)
	add = incr.WithinBindScope(scope, add)
	rem = &diffMapByKeysRemovedIncr[K, V]{
		n: incr.NewNode("diff_maps_by_keys_removed"),
		i: i,
	}
	incr.Link(rem, i)
	rem = incr.WithinBindScope(scope, rem)
	return
}

// DiffMapByKeysAdded returns an incremental that takes an input map typed
// incremental, and each stabilization pass returns just the subset
// of the map that was added since the last pass according to the keys.
func DiffMapByKeysAdded[K comparable, V any](scope *incr.BindScope, i incr.Incr[map[K]V]) incr.Incr[map[K]V] {
	o := &diffMapByKeysAddedIncr[K, V]{
		n: incr.NewNode("diff_maps_by_keys_added"),
		i: i,
	}
	incr.Link(o, i)
	return incr.WithinBindScope(scope, o)
}

// DiffMapByKeysRemoved returns an incremental that takes an input map typed
// incremental, and each stabilization pass returns just the subset
// of the map that was removed since the last pass according to the keys.
func DiffMapByKeysRemoved[K comparable, V any](scope *incr.BindScope, i incr.Incr[map[K]V]) incr.Incr[map[K]V] {
	o := &diffMapByKeysRemovedIncr[K, V]{
		n: incr.NewNode("diff_maps_by_keys_removed"),
		i: i,
	}
	incr.Link(o, i)
	return incr.WithinBindScope(scope, o)
}

var (
	_ incr.Incr[map[string]int] = (*diffMapByKeysAddedIncr[string, int])(nil)
	_ incr.INode                = (*diffMapByKeysAddedIncr[string, int])(nil)
	_ incr.IStabilize           = (*diffMapByKeysAddedIncr[string, int])(nil)
	_ fmt.Stringer              = (*diffMapByKeysAddedIncr[string, int])(nil)
)

type diffMapByKeysAddedIncr[K comparable, V any] struct {
	n    *incr.Node
	i    incr.Incr[map[K]V]
	last map[K]V
	val  map[K]V
}

func (mfn *diffMapByKeysAddedIncr[K, V]) String() string {
	return mfn.n.String()
}

func (mfn *diffMapByKeysAddedIncr[K, V]) Node() *incr.Node { return mfn.n }

func (mfn *diffMapByKeysAddedIncr[K, V]) Value() map[K]V { return mfn.val }

func (mfn *diffMapByKeysAddedIncr[K, V]) Stabilize(_ context.Context) error {
	newVal := mfn.i.Value()
	mfn.val, mfn.last = diffMapByKeysAdded(mfn.last, newVal)
	return nil
}

var (
	_ incr.Incr[map[string]int] = (*diffMapByKeysRemovedIncr[string, int])(nil)
	_ incr.INode                = (*diffMapByKeysRemovedIncr[string, int])(nil)
	_ incr.IStabilize           = (*diffMapByKeysRemovedIncr[string, int])(nil)
	_ fmt.Stringer              = (*diffMapByKeysRemovedIncr[string, int])(nil)
)

type diffMapByKeysRemovedIncr[K comparable, V any] struct {
	n    *incr.Node
	i    incr.Incr[map[K]V]
	last map[K]V
	val  map[K]V
}

func (mfn *diffMapByKeysRemovedIncr[K, V]) String() string {
	return mfn.n.String()
}

func (mfn *diffMapByKeysRemovedIncr[K, V]) Node() *incr.Node { return mfn.n }

func (mfn *diffMapByKeysRemovedIncr[K, V]) Value() map[K]V { return mfn.val }

func (mfn *diffMapByKeysRemovedIncr[K, V]) Stabilize(_ context.Context) error {
	newVal := mfn.i.Value()
	mfn.val, mfn.last = diffMapByKeysRemoved(mfn.last, newVal)
	return nil
}

func diffMapByKeysAdded[K comparable, V any](m0, m1 map[K]V) (add, orig map[K]V) {
	add = make(map[K]V)
	orig = make(map[K]V)
	var ok bool
	if len(m0) > 0 {
		for k, v := range m1 {
			if _, ok = m0[k]; !ok {
				add[k] = v
			}
			orig[k] = v
		}
		return
	}
	for k, v := range m1 {
		add[k] = v
		orig[k] = v
	}
	return
}

func diffMapByKeysRemoved[K comparable, V any](m0, m1 map[K]V) (rem, orig map[K]V) {
	rem = make(map[K]V)
	orig = make(map[K]V)
	var ok bool
	if len(m1) > 0 {
		for k, v := range m0 {
			if _, ok = m1[k]; !ok {
				rem[k] = v
			}
		}
		for k, v := range m1 {
			orig[k] = v
		}
		return
	}
	for k, v := range m0 {
		rem[k] = v
	}
	for k, v := range m1 {
		orig[k] = v
	}
	return
}
