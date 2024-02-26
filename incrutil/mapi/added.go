package mapi

import (
	"context"
	"maps"

	"github.com/wcharczuk/go-incr"
)

// Added returns an incremental node whose value is just the added keys (and their associated values)
// of an input map between stabilizations.
func Added[M ~map[K]V, K comparable, V any](scope incr.Scope, i incr.Incr[M]) incr.Incr[M] {
	return incr.WithinScope(scope, &addedIncr[M, K, V]{
		n:       incr.NewNode("mapi_added"),
		i:       i,
		parents: []incr.INode{i},
	})
}

type addedIncr[M ~map[K]V, K comparable, V any] struct {
	n       *incr.Node
	i       incr.Incr[M]
	parents []incr.INode
	last    M
	val     M
}

func (mfn *addedIncr[M, K, V]) Parents() []incr.INode {
	return mfn.parents
}

func (mfn *addedIncr[M, K, V]) String() string {
	return mfn.n.String()
}

func (mfn *addedIncr[M, K, V]) Node() *incr.Node { return mfn.n }

func (mfn *addedIncr[M, K, V]) Value() M { return mfn.val }

func (mfn *addedIncr[M, K, V]) Stabilize(_ context.Context) error {
	newVal := mfn.i.Value()
	mfn.val = symmetricDiffAdded[M, K, V](mfn.last, newVal)
	mfn.last = maps.Clone(newVal)
	return nil
}
