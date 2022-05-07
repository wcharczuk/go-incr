package incr

func newRecomputeHeap(maxHeight int) *recomputeHeap {
	return &recomputeHeap{
		maxHeight: maxHeight,
		heights:   make([][]GraphNode, maxHeight),
	}
}

// recomputeHeap is a height ordered set of nodes.
type recomputeHeap struct {
	maxHeight int
	heights   [][]GraphNode
	len       int
}

// add adds a node to the recompute heap at a given height.
func (rh *recomputeHeap) add(s GraphNode) {
	sn := s.Node()
	if sn.height >= rh.maxHeight {
		panic("recompute heap; cannot add node with height greater than max height")
	}
	rh.heights[sn.height] = append(rh.heights[sn.height], s)
	rh.len++
}

func (rh *recomputeHeap) has(s GraphNode) bool {
	sn := s.Node()
	if sn.height >= rh.maxHeight {
		panic("recompute heap; cannot has node with height greater than max height")
	}
	for _, n := range rh.heights[sn.height] {
		if n.Node().id == s.Node().id {
			return true
		}
	}
	return false
}

// removeMin removes the minimum node from the recompute heap.
func (rh *recomputeHeap) removeMin() GraphNode {
	for height := range rh.heights {
		if len(rh.heights[height]) > 0 {
			var first GraphNode
			first, rh.heights[height] = rh.removeFirst(rh.heights[height])
			rh.len--
			return first
		}
	}
	return nil
}

func (rh *recomputeHeap) removeFirst(stack []GraphNode) (first GraphNode, rest []GraphNode) {
	first = stack[0]
	rest = stack[1:]
	return
}
