package incr

import (
	"context"
	"fmt"
)

// ApplyN applies a function to given list of input incrementals and returns
// a new incremental of the output type of that function.
func ApplyN[A, B any](fn ApplyNFunc[A, B], inputs ...Incr[A]) ApplyNIncr[A, B] {
	o := &applyNIncr[A, B]{
		n:      NewNode(),
		inputs: inputs,
		fn:     fn,
	}
	for _, i := range inputs {
		Link(o, i)
	}
	return o
}

// ApplyNFunc is the function that the ApplyN incremental applies.
type ApplyNFunc[A, B any] func(context.Context, ...A) (B, error)

// ApplyNIncr is a type of incremental that can add inputs over time.
type ApplyNIncr[A, B any] interface {
	Incr[B]
	AddInput(Incr[A])
}

var (
	_ Incr[string]            = (*applyNIncr[int, string])(nil)
	_ ApplyNIncr[int, string] = (*applyNIncr[int, string])(nil)
	_ INode                   = (*applyNIncr[int, string])(nil)
	_ IStabilize              = (*applyNIncr[int, string])(nil)
	_ fmt.Stringer            = (*applyNIncr[int, string])(nil)
)

type applyNIncr[A, B any] struct {
	n      *Node
	inputs []Incr[A]
	fn     ApplyNFunc[A, B]
	val    B
}

func (mn *applyNIncr[A, B]) AddInput(i Incr[A]) {
	mn.inputs = append(mn.inputs, i)
	Link(mn, i)
}

func (mn *applyNIncr[A, B]) Node() *Node { return mn.n }

func (mn *applyNIncr[A, B]) Value() B { return mn.val }

func (mn *applyNIncr[A, B]) Stabilize(ctx context.Context) (err error) {
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

func (mn *applyNIncr[A, B]) String() string {
	return FormatNode(mn.n, "apply_n")
}
