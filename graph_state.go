package incr

func newGraphState() *graphState {
	return &graphState{
		generation: 1,
		nodeLookup: make(map[nodeID]Stabilizer),
	}
}

type graphState struct {
	recomputeHeap []Stabilizer
	generation    uint64
	nodeLookup    map[nodeID]Stabilizer
}

func (gs *graphState) addNode(s Stabilizer) {
	gs.recomputeHeap = gs.insertSorted(gs.recomputeHeap, s)
	gs.nodeLookup[s.Node().id] = s
}

func (gs *graphState) insertSorted(rh []Stabilizer, s Stabilizer) []Stabilizer {
	insertAt := gs.insertSearch(rh, s)
	rh = append(rh, s)
	copy(rh[insertAt+1:], rh[insertAt:])
	rh[insertAt] = s
	return rh
}

func (gs *graphState) insertSearch(rh []Stabilizer, s Stabilizer) int {
	i, j := 0, len(rh)
	for i < j {
		h := int(uint(i+j) >> 1)
		if rh[h].Node().height < s.Node().height {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}
