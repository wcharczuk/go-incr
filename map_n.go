package incr

import (
	"context"
	"fmt"
)

// MapN applies a function to given list of input incrementals and returns
// a new incremental of the output type of that function.
func MapN[A, B any](scope Scope, fn MapNFunc[A, B], inputs ...Incr[A]) MapNIncr[A, B] {
	return MapNContext(scope, func(_ context.Context, i ...A) (B, error) {
		return fn(i...), nil
	}, inputs...)
}

// MapNContext applies a function to given list of input incrementals and returns
// a new incremental of the output type of that function.
func MapNContext[A, B any](scope Scope, fn MapNContextFunc[A, B], inputs ...Incr[A]) MapNIncr[A, B] {
	return WithinScope(scope, &mapNIncr[A, B]{
		n:      NewNode("map_n"),
		inputs: inputs,
		fn:     fn,
	})
}

// MapNFunc is the function that the MapN incremental applies.
type MapNFunc[A, B any] func(...A) B

// MapNContextFunc is the function that the MapNContext incremental applies.
type MapNContextFunc[A, B any] func(context.Context, ...A) (B, error)

// MapNIncr is a type of incremental that can add inputs over time.
type MapNIncr[A, B any] interface {
	Incr[B]
	AddInput(Incr[A]) error
}

var (
	_ Incr[string]          = (*mapNIncr[int, string])(nil)
	_ MapNIncr[int, string] = (*mapNIncr[int, string])(nil)
	_ INode                 = (*mapNIncr[int, string])(nil)
	_ IStabilize            = (*mapNIncr[int, string])(nil)
	_ fmt.Stringer          = (*mapNIncr[int, string])(nil)
)

type mapNIncr[A, B any] struct {
	n      *Node
	inputs []Incr[A]
	fn     MapNContextFunc[A, B]
	val    B
}

func (mi *mapNIncr[A, B]) Parents() []INode {
	output := make([]INode, len(mi.inputs))
	for i := 0; i < len(mi.inputs); i++ {
		output[i] = mi.inputs[i]
	}
	return output
}

func (mn *mapNIncr[A, B]) AddInput(i Incr[A]) error {
	mn.inputs = append(mn.inputs, i)
	if mn.n.height != HeightUnset {
		// if we're already part of the graph, we have
		// to tell the graph to update our parent<>child metadata
		return GraphForNode(mn).addChild(mn, i)
	}
	return nil
}

func (mn *mapNIncr[A, B]) Node() *Node { return mn.n }

func (mn *mapNIncr[A, B]) Value() B { return mn.val }

func (mn *mapNIncr[A, B]) Stabilize(ctx context.Context) (err error) {
	var val B
	values := make([]A, len(mn.inputs))
	for index := range mn.inputs {
		values[index] = mn.inputs[index].Value()
	}
	val, err = mn.fn(ctx, values...)
	if err != nil {
		return
	}
	mn.val = val
	return nil
}

func (mn *mapNIncr[A, B]) String() string {
	return mn.n.String()
}
