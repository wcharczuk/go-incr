package incr

// Observe observes a node, specifically including it for computation
// as well as all of its parents.
func Observe[A any](g *Graph, input Incr[A]) ObserveIncr[A] {
	return nil
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

type observeIncr[A any] struct {
	n     *Node
	input Incr[A]
	value A
}
