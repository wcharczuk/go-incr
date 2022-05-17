package incr

import "sync"

// defaultRecomputeHeapMaxHeight is the default
// maximum recompute heap height when we create graph states.
const defaultRecomputeHeapMaxHeight = 255

// newGraphState returns a new graph state, which is the type that
// represents the shared state of a computation graph.
func newGraph() *graph {
	return &graph{
		id:               NewIdentifier(),
		stabilizationNum: 1,
		status:           StatusNotStabilizing,
		recomputeHeap:    newRecomputeHeap(defaultRecomputeHeapMaxHeight),
	}
}

type graph struct {
	// id is a unique identifier for the graph
	id Identifier
	// mu is a synchronizing mutex for the graph
	mu sync.Mutex
	// recomputeHeap is the heap of nodes to be processed
	// organized by pseudo-height
	recomputeHeap *recomputeHeap
	// setDuringStabilization is a list of nodes that were
	// set during stabilization
	setDuringStabilization *list[Identifier, INode]
	// stabilizationNum is the version
	// of the graph in respect to when
	// nodes are considered stale or changed
	stabilizationNum uint64
	// status is the general status of the graph where
	// the possible states are:
	// - StatusNotStabilizing (default)
	// - StatusStabilizing
	// - StatusRunningUpdateHandlers
	status int32
	// numNodes are the total number of nodes found during
	// discovery and is typically used for testing
	numNodes uint64
	// numNodesRecomputed is the total number of nodes
	// that have been recomputed in the graph's history
	// and is typically used in testing
	numNodesRecomputed uint64
	// numNodesChanged is the total number of nodes
	// that have been changed in the graph's history
	// and is typically used in testing
	numNodesChanged uint64
}
