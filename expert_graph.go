package incr

// ExpertGraph returns an "expert" interface to modify
// internal fields of the graph type.
//
// USE AT YOUR OWN RISK.
func ExpertGraph(g *Graph) IExpertGraph {
	return &expertGraph{graph: g}
}

// IExpertGraph is an interface to allow you to manage
// internal fields of a graph (this is useful if you're
// deserializing the graph from a durable store).
type IExpertGraph interface {
	SetID(Identifier)
	NumNodes() uint64
	NumNodesRecomputed() uint64
	NumNodesChanged() uint64
	StabilizationNum() uint64
	SetStabilizationNum(uint64)
	AddRecomputeHeap(...INode)
	RecomputeHeapLen() int
	RecomputeHeap() []Identifier
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

func (eg *expertGraph) AddRecomputeHeap(nodes ...INode) {
	eg.graph.recomputeHeap.Add(nodes...)
}

func (eg *expertGraph) RecomputeHeapLen() int {
	return eg.graph.recomputeHeap.Len()
}

func (eg *expertGraph) RecomputeHeap() []Identifier {
	output := make([]Identifier, 0, len(eg.graph.recomputeHeap.lookup))
	for _, n := range eg.graph.recomputeHeap.lookup {
		output = append(output, n.key)
	}
	return output
}
