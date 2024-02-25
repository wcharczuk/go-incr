package mapi

import (
	"context"
	"maps"

	"github.com/wcharczuk/go-incr"
)

func Removed[M ~map[K]V, K comparable, V any](scope incr.Scope, i incr.Incr[M]) incr.Incr[M] {
	return incr.WithinScope(scope, &removedIncr[M, K, V]{
		n: incr.NewNode("mapi_removed"),
		i: i,
	})
}

type removedIncr[M ~map[K]V, K comparable, V any] struct {
	n    *incr.Node
	i    incr.Incr[M]
	last M
	val  M
}

func (mfn *removedIncr[M, K, V]) Parents() []incr.INode {
	return []incr.INode{mfn.i}
}

func (mfn *removedIncr[M, K, V]) String() string {
	return mfn.n.String()
}

func (mfn *removedIncr[M, K, V]) Node() *incr.Node { return mfn.n }

func (mfn *removedIncr[M, K, V]) Value() M { return mfn.val }

func (mfn *removedIncr[M, K, V]) Stabilize(_ context.Context) error {
	newVal := mfn.i.Value()
	mfn.val = symmetricDiffRemoved[M, K, V](mfn.last, newVal)
	mfn.last = maps.Clone(newVal)
	return nil
}
