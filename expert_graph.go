package incr

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
	NumObservers() uint64
	StabilizationNum() uint64
	SetStabilizationNum(uint64)

	RecomputeHeapAdd(...INode)
	RecomputeHeapLen() int
	RecomputeHeapIDs() []Identifier

	AddChild(INode, INode) error
	RemoveParent(INode, INode)

	ObserveNode(IObserver, INode) error
	UnobserveNode(IObserver, INode)
}

type expertGraph struct {
	graph *Graph
}

func (eg *expertGraph) NumNodes() uint64 {
	return eg.graph.numNodes
}

func (eg *expertGraph) NumObservers() uint64 {
	return uint64(len(eg.graph.observers))
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
	for _, n := range nodes {
		eg.graph.recomputeHeap.add(n)
	}
}

func (eg *expertGraph) RecomputeHeapLen() int {
	return eg.graph.recomputeHeap.len()
}

func (eg *expertGraph) RecomputeHeapIDs() []Identifier {
	output := make([]Identifier, 0, eg.graph.recomputeHeap.numItems)
	for _, height := range eg.graph.recomputeHeap.heights {
		if height != nil {
			for key := range height.items {
				output = append(output, key)
			}
		}
	}
	return output
}

func (eg *expertGraph) AddChild(child, parent INode) error {
	return eg.graph.addChild(child, parent)
}

func (eg *expertGraph) RemoveParent(child, parent INode) {
	eg.graph.removeParent(child, parent)
}

func (eg *expertGraph) ObserveNode(obs IObserver, node INode) error {
	return eg.graph.observeNode(obs, node)
}

func (eg *expertGraph) UnobserveNode(obs IObserver, node INode) {
	eg.graph.unobserveNode(obs, node)
}
