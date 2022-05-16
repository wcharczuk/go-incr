package incr

// NodeStats return node statistics from a given node.
func NodeStats(n INode) INodeStats {
	return nodeStats{
		numRecomputes: n.Node().numRecomputes,
		numChanges:    n.Node().numChanges,
		numParents:    len(n.Node().parents),
		numChildren:   len(n.Node().children),
	}
}

// GraphStats return graph statistics from a given node.
func GraphStats(n INode) IGraphStats {
	return graphStats{
		numNodes:           n.Node().g.numNodes,
		numNodesRecomputed: n.Node().g.numNodesRecomputed,
		numNodesChanged:    n.Node().g.numNodesChanged,
	}
}

// INodeStats are stats for a given node.
type INodeStats interface {
	Recomputes() uint64
	Changes() uint64
	Children() int
	Parents() int
}

// IGraphStats are stats for a given node's graph.
type IGraphStats interface {
	Nodes() uint64
	NodesRecomputed() uint64
	NodesChanged() uint64
}

type nodeStats struct {
	numRecomputes uint64
	numChanges    uint64
	numChildren   int
	numParents    int
}

func (n nodeStats) Recomputes() uint64 { return n.numRecomputes }
func (n nodeStats) Changes() uint64    { return n.numChanges }
func (n nodeStats) Children() int      { return n.numChildren }
func (n nodeStats) Parents() int       { return n.numParents }

type graphStats struct {
	numNodes           uint64
	numNodesRecomputed uint64
	numNodesChanged    uint64
}

func (g graphStats) Nodes() uint64           { return g.numNodes }
func (g graphStats) NodesRecomputed() uint64 { return g.numNodesRecomputed }
func (g graphStats) NodesChanged() uint64    { return g.numNodesChanged }
