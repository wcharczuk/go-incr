package incr

const defaultRecomputeHeapMaxHeight = 255

func newGraphState() *graphState {
	return &graphState{
		id: NewIdentifier(),
		rh: newRecomputeHeap(defaultRecomputeHeapMaxHeight),
	}
}

type graphState struct {
	id Identifier
	sn uint64
	rh *recomputeHeap
	s  Status
}
