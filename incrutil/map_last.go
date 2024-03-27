package incrutil

import (
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
)

// MapLast returns an incremental that calls a given function with the previous value of
// an input incremental and a current value of an input incremental.
func MapLast[A, B any](scope incr.Scope, from incr.Incr[A], fn func(A, A) B) incr.Incr[B] {
	ml := &mapLastIncr[A, B]{
		n:  incr.NewNode("map_last"),
		i:  from,
		fn: fn,
	}
	incr.WithinScope(scope, ml)
	return ml
}

var (
	_ incr.Incr[any] = (*mapLastIncr[any, any])(nil)
	_ incr.IParents  = (*mapLastIncr[any, any])(nil)
	_ fmt.Stringer   = (*mapLastIncr[any, any])(nil)
)

type mapLastIncr[A, B any] struct {
	n     *incr.Node
	i     incr.Incr[A]
	fn    func(previous A, newValue A) B
	last  A
	value B
}

func (ml *mapLastIncr[A, B]) Parents() []incr.INode { return []incr.INode{ml.i} }

func (ml *mapLastIncr[A, B]) Node() *incr.Node { return ml.n }

func (ml *mapLastIncr[A, B]) Value() B { return ml.value }

func (ml *mapLastIncr[A, B]) Stabilize(ctx context.Context) error {
	current := ml.i.Value()
	ml.value = ml.fn(ml.last, current)
	ml.last = current
	return nil
}

func (ml *mapLastIncr[A, B]) String() string { return ml.n.String() }
