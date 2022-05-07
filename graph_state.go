package incr

const defaultRecomputeHeapMaxHeight = 255

func newGraphState() *graphState {
	return &graphState{
		id: newIdentifier(),
		rh: newRecomputeHeap(defaultRecomputeHeapMaxHeight),
	}
}

type graphState struct {
	id identifier
	sn uint64
	rh *recomputeHeap
	s  Status
}
