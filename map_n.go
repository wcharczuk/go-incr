package incr

import (
	"context"
	"fmt"
)

// MapN applies a function to given list of input incrementals and returns
// a new incremental of the output type of that function.
func MapN[A, B any](ctx context.Context, fn MapNFunc[A, B], inputs ...Incr[A]) MapNIncr[A, B] {
	return MapNContext(ctx, func(_ context.Context, i ...A) (B, error) {
		return fn(i...), nil
	}, inputs...)
}

// MapNContext applies a function to given list of input incrementals and returns
// a new incremental of the output type of that function.
func MapNContext[A, B any](ctx context.Context, fn MapNContextFunc[A, B], inputs ...Incr[A]) MapNIncr[A, B] {
	o := &mapNIncr[A, B]{
		n:      NewNode(),
		inputs: inputs,
		fn:     fn,
	}
	for _, i := range inputs {
		Link(o, i)
	}
	return WithinBindScope(ctx, o)
}

// MapNFunc is the function that the ApplyN incremental applies.
type MapNFunc[A, B any] func(...A) B

// MapNContextFunc is the function that the ApplyN incremental applies.
type MapNContextFunc[A, B any] func(context.Context, ...A) (B, error)

// MapNIncr is a type of incremental that can add inputs over time.
type MapNIncr[A, B any] interface {
	Incr[B]
	AddInput(Incr[A])
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

func (mn *mapNIncr[A, B]) AddInput(i Incr[A]) {
	mn.inputs = append(mn.inputs, i)
	Link(mn, i)
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
	return mn.n.String("map_n")
}
