package mapi

import (
	"cmp"
	"context"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/incrutil/pmap"
)

// Partition splits an incremental map in two by a predicate, recomputing only the entries
// that changed.
//
// A key whose predicate result flips moves from one side to the other, so this tracks the
// predicate as well as the values. Deriving the two halves with two [FilterMapValues]
// nodes would work and cost the same asymptotically, but would evaluate the predicate
// twice per changed key; here it is evaluated once.
func Partition[K cmp.Ordered, V any](
	scope incr.Scope,
	input incr.Incr[pmap.Map[K, V]],
	equal func(a, b V) bool,
	predicate func(K, V) bool,
) incr.Incr[Halves[K, V]] {
	p := &partitionIncr[K, V]{
		n:         incr.NewNode("mapi_partition"),
		i:         input,
		equal:     equal,
		predicate: predicate,
	}
	p.parents[0] = input
	return incr.WithinScope(scope, p)
}

// Halves is the result of a [Partition]: the entries the predicate accepted and those it
// did not.
type Halves[K cmp.Ordered, V any] struct {
	Matching    pmap.Map[K, V]
	NotMatching pmap.Map[K, V]
}

var (
	_ incr.Incr[Halves[string, int]] = (*partitionIncr[string, int])(nil)
	_ incr.IStabilize                = (*partitionIncr[string, int])(nil)
	_ incr.IParents                  = (*partitionIncr[string, int])(nil)
)

type partitionIncr[K cmp.Ordered, V any] struct {
	n         *incr.Node
	i         incr.Incr[pmap.Map[K, V]]
	equal     func(a, b V) bool
	predicate func(K, V) bool
	last      pmap.Map[K, V]
	value     Halves[K, V]
	parents   [1]incr.INode
}

func (p *partitionIncr[K, V]) Parents() []incr.INode { return p.parents[:] }

func (p *partitionIncr[K, V]) Node() *incr.Node { return p.n }

func (p *partitionIncr[K, V]) Value() Halves[K, V] { return p.value }

func (p *partitionIncr[K, V]) Stabilize(_ context.Context) error {
	current := p.i.Value()
	out := p.value
	for change := range p.last.SymmetricDiff(current, p.equal) {
		if change.Kind == pmap.ChangeRemoved {
			out.Matching = out.Matching.Delete(change.Key)
			out.NotMatching = out.NotMatching.Delete(change.Key)
			continue
		}
		// a key can be on either side already, and the predicate may have flipped, so
		// place it on one side and make sure it is absent from the other
		if p.predicate(change.Key, change.New) {
			out.Matching = out.Matching.Set(change.Key, change.New)
			out.NotMatching = out.NotMatching.Delete(change.Key)
		} else {
			out.NotMatching = out.NotMatching.Set(change.Key, change.New)
			out.Matching = out.Matching.Delete(change.Key)
		}
	}
	p.value = out
	p.last = current
	return nil
}

func (p *partitionIncr[K, V]) String() string { return p.n.String() }
