package incr

import (
	"context"
	"fmt"
)

// MustObserve observes a node, specifically including it for computation
// as well as all of its parents.
//
// If this detects a cycle or any other issue a panic will be raised.
func MustObserve[A any](g *Graph, input Incr[A]) ObserveIncr[A] {
	o, err := Observe[A](g, input)
	if err != nil {
		panic(err)
	}
	return o
}

// Observe observes a node, specifically including it for computation
// as well as all of its parents.
func Observe[A any](g *Graph, input Incr[A]) (ObserveIncr[A], error) {
	o := WithinScope(g, &observeIncr[A]{
		n:     NewNode("observer"),
		input: input,
	})
	g.addObserver(o)
	if err := g.addChildObserver(o, input); err != nil {
		return nil, err
	}
	return o, nil
}

// ObserveIncr is an incremental that observes a graph
// of incrementals starting a given input.
type ObserveIncr[A any] interface {
	INode
	// Value returns the observed node value.
	Value() A
	// Unobserve effectively removes a given node from the observed ref count for a graph.
	//
	// As well, it unlinks the observer from its parent nodes, and as a result
	// you should _not_ re-use the node.
	//
	// To observe parts of a graph again, use the `MustObserve(...)` helper.
	Unobserve(context.Context)
}

// IObserver is an INode that can be unobserved.
type IObserver interface {
	INode
	Unobserve(context.Context)
}

var (
	_ ObserveIncr[any] = (*observeIncr[any])(nil)
	_ INode            = (*observeIncr[any])(nil)
	_ fmt.Stringer     = (*observeIncr[any])(nil)
)

type observeIncr[A any] struct {
	n     *Node
	input Incr[A]
	value A
}

func (o *observeIncr[A]) Node() *Node { return o.n }

// Unobserve effectively removes a given node from the observed ref count for a graph.
//
// As well, it unlinks the observer from its parent nodes, and as a result
// you should _not_ re-use the node.
//
// To observe parts of a graph again, use the `MustObserve(...)` helper.
func (o *observeIncr[A]) Unobserve(ctx context.Context) {
	g := graphFromCreatedIn(o)
	g.removeObserver(o)
	o.input.Node().removeObserver(o.n.id)
	g.checkIfUnnecessary(o.input)
	var value A
	o.value = value
	o.input = nil
}

func (o *observeIncr[A]) Value() (output A) {
	if o.input == nil {
		return
	}
	return o.input.Value()
}

func (o *observeIncr[A]) String() string {
	if o.n.label != "" {
		return fmt.Sprintf("%s[%s]:%s", o.n.kind, o.n.id.Short(), o.n.label)
	}
	return fmt.Sprintf("%s[%s]", o.n.kind, o.n.id.Short())
}
