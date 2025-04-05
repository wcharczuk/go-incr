package incr

import "fmt"

// Always returns an incremental that is always stale and whose
// children will always be marked for recomputation.
//
// The [Always] node will pass through the input incremental value
// to the nodes that take the [Always] node as an input.
//
// An example use case for [Always] might be something like:
//
//	g := incr.New()
//	v := incr.Var(g, "hello")
//	mv := incr.Map(g, v, func(vv string) string { return "not-" + v })
//	a := incr.Always(g, mv)
//	var counter int
//	mva := incr.Map(g, a, func(vv string) string { counter++; return vv })
//	_ = g.Stabilize(context.TODO())
//
// In the above example, each time we call [Graph.Stabilize] we'll increment the counter.
func Always[A any](scope Scope, input Incr[A]) Incr[A] {
	return WithinScope(scope, &alwaysIncr[A]{
		n:       NewNode(KindAlways),
		input:   input,
		parents: []INode{input},
	})
}

var (
	_ Incr[any]    = (*alwaysIncr[any])(nil)
	_ IAlways      = (*alwaysIncr[any])(nil)
	_ IStale       = (*alwaysIncr[any])(nil)
	_ fmt.Stringer = (*alwaysIncr[any])(nil)
)

type alwaysIncr[A any] struct {
	n       *Node
	input   Incr[A]
	parents []INode
}

func (a *alwaysIncr[A]) Parents() []INode {
	return a.parents
}

func (a *alwaysIncr[A]) Stale() bool {
	return true
}

func (a *alwaysIncr[A]) Always() {}

func (a *alwaysIncr[A]) Value() A {
	return a.input.Value()
}

func (a *alwaysIncr[A]) Node() *Node { return a.n }

func (a *alwaysIncr[A]) String() string {
	return a.n.String()
}
