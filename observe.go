package incr

import (
	"context"
	"fmt"
)

// Observe observes a node, specifically including it for computation
// as well as all of its parents.
func Observe[A any](g *Graph, input Incr[A]) ObserveIncr[A] {
	o := WithinScope(g, &observeIncr[A]{
		n:       NewNode("observer"),
		input:   input,
		parents: []INode{input},
	})
	g.addNodeOrObserver(o)
	_ = g.addNewObserverToNode(o, input)
	if input.Node().height >= o.Node().height {
		_ = g.adjustHeightsHeap.setHeight(o, input.Node().height+1)
		_ = g.adjustHeightsHeap.adjustHeights(g.recomputeHeap, o, input)
	}
	return o
}

// ObserveIncr is an incremental that observes a graph
// of incrementals starting a given input.
type ObserveIncr[A any] interface {
	Incr[A]
	// Unobserve effectively removes a given node from the observed ref count for a graph.
	//
	// As well, it unlinks the observer from its parent nodes, and as a result
	// you should _not_ re-use the node.
	//
	// To observe parts of a graph again, use the `Observe(...)` helper.
	Unobserve(context.Context)
}

// IObserver is an INode that can be unobserved.
type IObserver interface {
	INode
	Unobserve(context.Context)
}

var (
	_ Incr[any]    = (*observeIncr[any])(nil)
	_ IStabilize   = (*observeIncr[any])(nil)
	_ INode        = (*observeIncr[any])(nil)
	_ fmt.Stringer = (*observeIncr[any])(nil)
)

type observeIncr[A any] struct {
	n       *Node
	input   Incr[A]
	value   A
	parents []INode
}

func (o *observeIncr[A]) Parents() []INode {
	return o.parents
}

func (o *observeIncr[A]) Node() *Node { return o.n }

func (o *observeIncr[A]) Stabilize(_ context.Context) error {
	o.value = o.input.Value()
	return nil
}

// Unobserve effectively removes a given node from the observed ref count for a graph.
//
// As well, it unlinks the observer from its parent nodes, and as a result
// you should _not_ re-use the node.
//
// To observe parts of a graph again, use the `Observe(...)` helper.
func (o *observeIncr[A]) Unobserve(ctx context.Context) {
	g := o.n.graph

	o.input.Node().removeObserver(o.n.id)
	// Unlink(o, o.input)
	g.removeObserver(o)

	// zero out the observed value
	var value A
	o.value = value
	o.input = nil
}

func (o *observeIncr[A]) Value() (output A) {
	return o.value
}

func (o *observeIncr[A]) String() string {
	return o.n.String()
}
