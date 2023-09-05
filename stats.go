package incr

// NodeStats return node statistics from a given node.
func NodeStats(n INode) INodeStats {
	return nodeStats{
		numRecomputes: n.Node().numRecomputes,
		numChanges:    n.Node().numChanges,
		numParents:    len(n.Node().parents),
		numChildren:   len(n.Node().children),
		setAt:         n.Node().setAt,
		changedAt:     n.Node().changedAt,
	}
}

// GraphStats return graph statistics from a given node.
func (graph *Graph) Stats() IGraphStats {
	var recomputeHeapLength int
	if graph.recomputeHeap != nil {
		recomputeHeapLength = graph.recomputeHeap.Len()
	}
	return graphStats{
		stabilizationNum:    graph.stabilizationNum,
		numNodes:            graph.numNodes,
		numNodesRecomputed:  graph.numNodesRecomputed,
		numNodesChanged:     graph.numNodesChanged,
		recomputeHeapLength: recomputeHeapLength,
	}
}

// INodeStats are stats for a given node.
type INodeStats interface {
	Recomputes() uint64
	Changes() uint64
	SetAt() uint64
	ChangedAt() uint64
	Children() int
	Parents() int
}

// IGraphStats are stats for a given node's graph.
type IGraphStats interface {
	StabilizationNum() uint64
	Nodes() uint64
	NodesRecomputed() uint64
	NodesChanged() uint64
	RecomputeHeapLength() int
}

type nodeStats struct {
	numRecomputes uint64
	numChanges    uint64
	numChildren   int
	numParents    int
	setAt         uint64
	changedAt     uint64
}

func (n nodeStats) Recomputes() uint64 { return n.numRecomputes }
func (n nodeStats) Changes() uint64    { return n.numChanges }
func (n nodeStats) Children() int      { return n.numChildren }
func (n nodeStats) Parents() int       { return n.numParents }
func (n nodeStats) SetAt() uint64      { return n.setAt }
func (n nodeStats) ChangedAt() uint64  { return n.changedAt }

type graphStats struct {
	stabilizationNum    uint64
	numNodes            uint64
	numNodesRecomputed  uint64
	numNodesChanged     uint64
	recomputeHeapLength int
}

func (g graphStats) StabilizationNum() uint64 { return g.stabilizationNum }
func (g graphStats) Nodes() uint64            { return g.numNodes }
func (g graphStats) NodesRecomputed() uint64  { return g.numNodesRecomputed }
func (g graphStats) NodesChanged() uint64     { return g.numNodesChanged }
func (g graphStats) RecomputeHeapLength() int { return g.recomputeHeapLength }
