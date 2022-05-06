package incr

func newGraphState() *graphState {
	return &graphState{
		nodeLookup: make(map[nodeID]Stabilizer),
	}
}

type graphState struct {
	recomputeHeap []Stabilizer
	generation    uint64
	nodeLookup    map[nodeID]Stabilizer
}

func (gs *graphState) addRecomputeHeap(s Stabilizer) {
	insertAt := gs.insertSearch(gs.recomputeHeap, s)
	gs.recomputeHeap = append(gs.recomputeHeap, s)
	copy(gs.recomputeHeap[insertAt+1:], gs.recomputeHeap[insertAt:])
	gs.recomputeHeap[insertAt] = s
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
