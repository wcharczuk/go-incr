package incr

import (
	"context"
	"fmt"
)

// MapN applies a function to given list of input incrementals and returns
// a new incremental of the output type of that function.
//
// The slice of values passed to fn is reused between stabilizations and is
// only valid for the duration of the call; do not retain it past fn returning.
func MapN[A, B any](scope Scope, fn MapNFunc[A, B], inputs ...Incr[A]) MapNIncr[A, B] {
	return MapNContext(scope, func(_ context.Context, i ...A) (B, error) {
		return fn(i...), nil
	}, inputs...)
}

// MapNContext applies a function to given list of input incrementals and returns
// a new incremental of the output type of that function.
//
// The slice of values passed to fn is reused between stabilizations and is
// only valid for the duration of the call; do not retain it past fn returning.
func MapNContext[A, B any](scope Scope, fn MapNContextFunc[A, B], inputs ...Incr[A]) MapNIncr[A, B] {
	m := &mapNIncr[A, B]{
		n:      scope.newNode(KindMapN),
		inputs: inputs,
		fn:     fn,
	}
	m.parents = make([]INode, len(inputs))
	for i := range inputs {
		m.parents[i] = inputs[i]
	}
	return WithinScope(scope, m)
}

// MapNFunc is the function that the MapN incremental applies.
type MapNFunc[A, B any] func(...A) B

// MapNContextFunc is the function that the MapNContext incremental applies.
type MapNContextFunc[A, B any] func(context.Context, ...A) (B, error)

// MapNIncr is a type of incremental that can add inputs over time.
type MapNIncr[A, B any] interface {
	Incr[B]
	IParents
	IStabilize
	AddInput(Incr[A]) error
	RemoveInput(Identifier) error
}

var (
	_ Incr[string]          = (*mapNIncr[int, string])(nil)
	_ MapNIncr[int, string] = (*mapNIncr[int, string])(nil)
	_ INode                 = (*mapNIncr[int, string])(nil)
	_ IStabilize            = (*mapNIncr[int, string])(nil)
	_ fmt.Stringer          = (*mapNIncr[int, string])(nil)
)

type mapNIncr[A, B any] struct {
	n       *Node
	inputs  []Incr[A]
	fn      MapNContextFunc[A, B]
	val     B
	parents []INode
	// values is a reused buffer of the inputs' values that is passed to fn
	// on each stabilization to avoid allocating a fresh slice every pass.
	values []A
}

func (mi *mapNIncr[A, B]) Parents() []INode {
	return mi.parents
}

func (mn *mapNIncr[A, B]) AddInput(i Incr[A]) error {
	mn.inputs = append(mn.inputs, i)
	mn.parents = append(mn.parents, i)
	if mn.n.height != HeightUnset {
		// if we're already part of the graph, we have
		// to tell the graph to update our parent<>child metadata
		return GraphForNode(mn).addChild(mn, i)
	}
	return nil
}

func (mn *mapNIncr[A, B]) RemoveInput(id Identifier) error {
	var removed Incr[A]
	mn.inputs, removed = remove(mn.inputs, id)
	if removed != nil {
		mn.parents, _ = remove(mn.parents, id)
		mn.Node().removeParent(id)
		removed.Node().removeChild(mn.n.id)
		GraphForNode(mn).SetStale(mn)
		GraphForNode(mn).checkIfUnnecessary(removed)
		return nil
	}
	return nil
}

func (mn *mapNIncr[A, B]) Node() *Node { return mn.n }

func (mn *mapNIncr[A, B]) Value() B { return mn.val }

func (mn *mapNIncr[A, B]) Stabilize(ctx context.Context) (err error) {
	// reuse the values buffer across stabilizations, growing it only when the
	// set of inputs grows. note that the slice handed to fn is owned by this
	// node and is only valid for the duration of the call; callers must not
	// retain it past their function returning.
	if cap(mn.values) < len(mn.inputs) {
		mn.values = make([]A, len(mn.inputs))
	} else {
		mn.values = mn.values[:len(mn.inputs)]
	}
	for index := range mn.inputs {
		mn.values[index] = mn.inputs[index].Value()
	}
	var val B
	val, err = mn.fn(ctx, mn.values...)
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *mapNIncr[A, B]) String() string {
	return mn.n.String()
}
