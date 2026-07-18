package mapi

import (
	"cmp"
	"context"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// Join turns a map of incrementals into an incremental map, reading each entry's
// current value.
//
// Each inner incremental is linked as an input while its key is in the map and
// unlinked when the key leaves, so an inner computation is only necessary for as long
// as something is actually reading it. That is the same dynamic linking
// [incr.MapNIncr.AddInput] performs, driven here by the diff of the outer map.
//
// Both kinds of change cost only what changed:
//
//   - a key added, removed, or repointed at a different incremental: one link or
//     unlink, O(changes x log n).
//   - an inner incremental producing a new value: the node is told which of its
//     inputs moved, so only that key's entry is rewritten.
//
// The equality used on the outer map is identity of the inner incrementals, since
// what matters there is whether a key now points at a different computation, not
// whether that computation's value has moved.
func Join[K cmp.Ordered, V any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, incr.Incr[V]]],
) incr.Incr[pmap.Map[K, V]] {
	j := &joinIncr[K, V]{
		n:      incr.NewNode("mapi_join"),
		i:      input,
		byNode: make(map[incr.Identifier]K),
	}
	j.parents = append(j.parents, input)
	return incr.WithinScope(scope, j)
}

var (
	_ incr.Incr[pmap.Map[string, int]] = (*joinIncr[string, int])(nil)
	_ incr.IStabilize                  = (*joinIncr[string, int])(nil)
	_ incr.IParents                    = (*joinIncr[string, int])(nil)
	_ incr.IChildChanged               = (*joinIncr[string, int])(nil)
)

type joinIncr[K cmp.Ordered, V any] struct {
	n *incr.Node
	i incr.Incr[pmap.Map[K, incr.Incr[V]]]
	// last is the outer map as of the previous pass, retained to diff against.
	last pmap.Map[K, incr.Incr[V]]
	// linked is the inner incremental currently linked for each key, so that a key
	// repointed at a different computation can unlink the old one.
	linked pmap.Map[K, incr.Incr[V]]
	// byNode maps a linked inner node back to its key, so a value change can be
	// attributed without searching.
	byNode  map[incr.Identifier]K
	value   pmap.Map[K, V]
	parents []incr.INode
	// pending holds the keys whose inner incremental reported a new value since the
	// last recompute.
	pending []K
}

func (j *joinIncr[K, V]) Parents() []incr.INode { return j.parents }

func (j *joinIncr[K, V]) Node() *incr.Node { return j.n }

func (j *joinIncr[K, V]) Value() pmap.Map[K, V] { return j.value }

// ChildChanged records that one of the linked inner incrementals took a new value.
func (j *joinIncr[K, V]) ChildChanged(child incr.INode) {
	if key, ok := j.byNode[child.Node().ID()]; ok {
		j.pending = append(j.pending, key)
	}
}

func (j *joinIncr[K, V]) Stabilize(_ context.Context) error {
	current := j.i.Value()
	out := j.value

	// structural changes first: the set of inner incrementals being read.
	for change := range j.last.SymmetricDiff(current, sameNode[V]) {
		switch change.Kind {
		case pmap.ChangeRemoved:
			j.unlink(change.Key)
			out = out.Delete(change.Key)
		case pmap.ChangeAdded:
			if err := j.link(change.Key, change.New); err != nil {
				return err
			}
			out = out.Set(change.Key, change.New.Value())
		case pmap.ChangeUpdated:
			// the key now points at a different computation
			j.unlink(change.Key)
			if err := j.link(change.Key, change.New); err != nil {
				return err
			}
			out = out.Set(change.Key, change.New.Value())
		}
	}

	// then value changes reported by the inner incrementals themselves.
	for _, key := range j.pending {
		inner, ok := j.linked.Get(key)
		if !ok {
			// the key was removed in this same pass
			continue
		}
		out = out.Set(key, inner.Value())
	}
	j.pending = j.pending[:0]

	j.value = out
	j.last = current
	return nil
}

// link makes an inner incremental an input of this node, so that it is necessary and
// its changes reach us.
func (j *joinIncr[K, V]) link(key K, inner incr.Incr[V]) error {
	j.linked = j.linked.Set(key, inner)
	j.byNode[inner.Node().ID()] = key
	j.parents = append(j.parents, inner)
	if incr.ExpertNode(j).Height() == incr.HeightUnset {
		// not yet part of the graph; the edge will be established when it is
		return nil
	}
	// already part of the graph, so the graph has to be told about the new edge and
	// given the chance to restore the height ordering.
	if err := incr.ExpertGraph(incr.GraphForNode(j)).AddChild(j, inner); err != nil {
		return err
	}
	// A newly linked inner incremental may never have been computed, because nothing
	// needed it until now, in which case its value is not yet meaningful. Linking
	// makes it necessary and schedules it, but this node has already recomputed in
	// this pass, so it will not look stale with respect to a parent that changes
	// during the same pass and would not run again to collect the value. Marking
	// ourselves stale schedules the second pass that does; heights guarantee it runs
	// after the inner node.
	incr.GraphForNode(j).SetStale(j)
	return nil
}

// unlink drops an inner incremental, releasing it if nothing else needs it.
func (j *joinIncr[K, V]) unlink(key K) {
	inner, ok := j.linked.Get(key)
	if !ok {
		return
	}
	j.linked = j.linked.Delete(key)
	delete(j.byNode, inner.Node().ID())
	for index, parent := range j.parents {
		if parent.Node().ID() == inner.Node().ID() {
			j.parents = append(j.parents[:index], j.parents[index+1:]...)
			break
		}
	}
	incr.ExpertGraph(incr.GraphForNode(j)).RemoveParent(j, inner)
}

// sameNode compares inner incrementals by identity rather than by value: the outer
// map has changed for a key when it points at a different computation.
func sameNode[V any](a, b incr.Incr[V]) bool {
	return a.Node().ID() == b.Node().ID()
}
