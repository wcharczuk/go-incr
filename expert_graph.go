package incr

// ExpertGraph returns an "expert" interface to modify
// internal fields of the graph type.
//
// Note there are no compatibility guarantees on this interface
// and you should use this interface at your own risk.
func ExpertGraph(g *Graph) IExpertGraph {
	return &expertGraph{graph: g}
}

// IExpertGraph is an interface to allow you to manage
// internal fields of a graph (this is useful if you're
// deserializing the graph from a durable store).
//
// Note there are no compatibility guarantees on this interface
// and you should use this interface at your own risk.
type IExpertGraph interface {
	// SetID sets the identifier for the [Graph].
	SetID(Identifier)

	// NumNodes returns the number of nodes the [Graph] is tracking.
	NumNodes() uint64

	// NumNodesRecomputed returns the number of nodes the [Graph] has
	// recomputed in its lifetime.
	NumNodesRecomputed() uint64

	// NumNodesChanged returns the number of nodes the [Graph] has
	// updated the value of in its lifetime.
	NumNodesChanged() uint64

	// NumObservers returns the current count of observers the [Graph] is tracking.
	NumObservers() uint64

	// StabilizationNum returns the current stabilization number of the [Graph].
	StabilizationNum() uint64

	// SetStabilizationNumber sets the current stabilization number, specifically
	// in situations where you're restoring graph state.
	SetStabilizationNum(uint64)

	// RecomputeHeapAdd directly adds a varadic array of nodes to the recompute heap.
	RecomputeHeapAdd(...INode)

	// RecomputeHeapLen returns the current length of the recompute heap.
	RecomputeHeapLen() int

	// RecomputeHeapIDs returns the node identifiers that are held in the recompute heap.
	//
	// This is useful when saving the state of a [Graph] to an external store.
	RecomputeHeapIDs() []Identifier

	// AddChild associates a child node to a parent.
	AddChild(child INode, parent INode) error
	// RemoveParent removes the association between a child and a parent.
	RemoveParent(child INode, parent INode)

	// ObserveNode implements the observe steps usually handled by [Observe] for custom nodes.
	ObserveNode(IObserver, INode) error

	// UnobserveNode implements the unobserve steps usually handled by observers.
	UnobserveNode(IObserver, INode)

	// CheckIfUnnecessary adds a node to the became unnecessary queue
	// if it is (newly) unnecessary.
	CheckIfUnnecessary(INode)
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
			cursor := height.head
			for cursor != nil {
				output = append(output, cursor.Node().id)
				cursor = cursor.Node().nextInRecomputeHeap
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

func (eg *expertGraph) CheckIfUnnecessary(node INode) {
	eg.graph.checkIfUnnecessary(node)
}
