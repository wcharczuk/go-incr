package incr

import "sync"

const defaultRecomputeHeapMaxHeight = 255

func newGraphState() *graphState {
	return &graphState{
		id: NewIdentifier(),
		sn: 1,
		rh: newRecomputeHeap(defaultRecomputeHeapMaxHeight),
	}
}

type graphState struct {
	id Identifier
	sn uint64
	rh *recomputeHeap
	s  Status
	mu sync.Mutex
}
