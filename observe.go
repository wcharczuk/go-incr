package incr

import "fmt"

// Observe observes a node, specifically including it for computation
// as well as all of its parents.
func Observe[A any](g *Graph, input Incr[A]) ObserveIncr[A] {
	o := &observeIncr[A]{
		n:     NewNode(),
		input: input,
	}
	Link(o, input)
	g.DiscoverObserver(o)
	g.DiscoverNodes(o, input)
	return o
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
	g.UndiscoverNode(o, o.input)
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
