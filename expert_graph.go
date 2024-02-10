package incr

import "context"

// ExpertGraph returns an "expert" interface to modify
// internal fields of the graph type.
//
// Note there are no compatibility guarantees on this interface
// and you should use this interface at your own caution.
func ExpertGraph(g *Graph) IExpertGraph {
	return &expertGraph{graph: g}
}

// IExpertGraph is an interface to allow you to manage
// internal fields of a graph (this is useful if you're
// deserializing the graph from a durable store).
//
// Note there are no compatibility guarantees on this interface
// and you should use this interface at your own caution.
type IExpertGraph interface {
	SetID(Identifier)
	NumNodes() uint64
	NumNodesRecomputed() uint64
	NumNodesChanged() uint64
	StabilizationNum() uint64
	SetStabilizationNum(uint64)

	RecomputeHeapAdd(...INode)
	RecomputeHeapLen() int
	RecomputeHeapIDs() []Identifier

	AddObserver(IObserver)
	RemoveObserver(IObserver)

	ObserveNodes(*BindScope, INode, ...IObserver)
	UnobserveNodes(context.Context, INode, ...IObserver)
}

type expertGraph struct {
	graph *Graph
}

func (eg *expertGraph) NumNodes() uint64 {
	return eg.graph.numNodes
}

func (eg *expertGraph) NumNodesRecomputed() uint64 {
	return eg.graph.numNodesRecomputed
}

func (eg *expertGraph) NumNodesChanged() uint64 {
	return eg.graph.numNodesChanged
}

func (eg *expertGraph) SetID(id Identifier) {
	eg.graph.id = id
}

func (eg *expertGraph) StabilizationNum() uint64 {
	return eg.graph.stabilizationNum
}

func (eg *expertGraph) SetStabilizationNum(stabilizationNum uint64) {
	eg.graph.stabilizationNum = stabilizationNum
}

func (eg *expertGraph) RecomputeHeapAdd(nodes ...INode) {
	eg.graph.recomputeHeap.add(nodes...)
}

func (eg *expertGraph) RecomputeHeapLen() int {
	return eg.graph.recomputeHeap.len()
}

func (eg *expertGraph) RecomputeHeapIDs() []Identifier {
	output := make([]Identifier, 0, len(eg.graph.recomputeHeap.lookup))
	for _, height := range eg.graph.recomputeHeap.heights {
		if height == nil {
			continue
		}
		for id := range height {
			output = append(output, id)
		}
	}
	return output
}

func (eg *expertGraph) ObserveNodes(scope *BindScope, n INode, observers ...IObserver) {
	eg.graph.observeNodes(scope, n, observers...)
}

func (eg *expertGraph) UnobserveNodes(ctx context.Context, n INode, observers ...IObserver) {
	eg.graph.unobserveNodes(ctx, n, observers...)
}

func (eg *expertGraph) AddObserver(on IObserver) {
	eg.graph.addObserver(on)
}

func (eg *expertGraph) RemoveObserver(on IObserver) {
	eg.graph.removeObserver(on)
}
