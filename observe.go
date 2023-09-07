package incr

import "fmt"

// MustObserve observes a node and panics if there is a cycle.
func MustObserve[A any](g *Graph, input Incr[A]) ObserveIncr[A] {
	o, err := Observe[A](g, input)
	if err != nil {
		panic(err)
	}
	return o
}

// Observe observes a node, specifically including it for computation
// as well as all of its parents.
func Observe[A any](g *Graph, input Incr[A]) (o ObserveIncr[A], err error) {
	o = &observeIncr[A]{
		n:     NewNode(),
		input: input,
	}
	Link(o, input)
	err = g.DiscoverObserver(o)
	if err != nil {
		return
	}
	err = g.DiscoverNodes(o, input)
	if err != nil {
		return
	}
	return
}

// ObserveIncr is an incremental that observes a graph
// of incrementals starting a given input.
type ObserveIncr[A any] interface {
	Incr[A]
	Unobserve()
}

// IObserver is an INode that can be unobserved.
type IObserver interface {
	INode
	Unobserve()
}

var (
	_ Incr[any]    = (*observeIncr[any])(nil)
	_ INode        = (*observeIncr[any])(nil)
	_ fmt.Stringer = (*observeIncr[any])(nil)
)

type observeIncr[A any] struct {
	n     *Node
	input Incr[A]
	value A
}

func (o *observeIncr[A]) Node() *Node { return o.n }

// Unobserve effectively removes a given node
// from the observed ref count for a graph.
func (o *observeIncr[A]) Unobserve() {
	g := o.n.graph
	g.UndiscoverNodes(o, o.input)
	g.UndiscoverObserver(o)
	o.input = nil
}

func (o *observeIncr[A]) Value() (output A) {
	if o.input != nil {
		output = o.input.Value()
	}
	return
}

func (o *observeIncr[A]) String() string {
	return o.n.String("observer")
}
