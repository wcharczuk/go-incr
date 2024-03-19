package incrutil

import (
	"context"
	"fmt"

	"github.com/wcharczuk/go-incr"
)

// Accumulate returns an incremental that accepts new values from an input incremental
// and returns an array of those values.
func Accumulate[A any](scope incr.Scope, from incr.Incr[A], fn func([]A, A) []A) incr.Incr[[]A] {
	ai := &accumulateIncr[A]{
		n:  incr.NewNode("accumulate"),
		i:  from,
		fn: fn,
	}
	incr.WithinScope(scope, ai)
	return ai
}

var (
	_ incr.Incr[[]any] = (*accumulateIncr[any])(nil)
	_ incr.IParents    = (*accumulateIncr[any])(nil)
	_ fmt.Stringer     = (*accumulateIncr[any])(nil)
)

type accumulateIncr[A any] struct {
	n     *incr.Node
	i     incr.Incr[A]
	fn    func(previous []A, newValue A) []A
	value []A
}

func (ai *accumulateIncr[A]) Parents() []incr.INode { return []incr.INode{ai.i} }

func (ai *accumulateIncr[A]) Node() *incr.Node { return ai.n }

func (ai *accumulateIncr[A]) Value() []A { return ai.value }

func (ai *accumulateIncr[A]) Stabilize(ctx context.Context) error {
	ai.value = ai.fn(ai.value, ai.i.Value())
	return nil
}

func (ai *accumulateIncr[A]) String() string { return ai.n.String() }
