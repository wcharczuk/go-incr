package incr

import (
	"fmt"
	"math"
	"sync"
)

func newRecomputeHeap(maxHeight int) *recomputeHeap {
	return &recomputeHeap{
		heights: make([]recomputeHeapList, maxHeight),
	}
}

type recomputeHeap struct {
	mu        sync.Mutex
	minHeight int
	maxHeight int
	// heights holds the nodes at each height inline rather than behind a
	// pointer per height, so scanning for the next non-empty block reads one
	// contiguous run of memory instead of chasing a pointer per probe.
	heights  []recomputeHeapList
	numItems int
}

func (rh *recomputeHeap) clear() (aborted []INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	var next INode
	aborted = make([]INode, 0, rh.numItems)
	for rh.numItems > 0 {
		next, _ = rh.removeMinUnsafe()
		next.Node().heightInRecomputeHeap = HeightUnset
		aborted = append(aborted, next)
	}

	rh.heights = make([]recomputeHeapList, len(rh.heights))
	rh.minHeight = 0
	rh.maxHeight = 0
	rh.numItems = 0
	return
}

func (rh *recomputeHeap) len() int {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	return rh.numItems
}

func (rh *recomputeHeap) add(nodes ...INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	for _, n := range nodes {
		rh.addNodeUnsafe(n)
	}
}

func (rh *recomputeHeap) addIfNotPresent(n INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	if n.Node().heightInRecomputeHeap == HeightUnset {
		rh.addNodeUnsafe(n)
	}
}

func (rh *recomputeHeap) fix(n INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.fixUnsafe(n)
}

func (rh *recomputeHeap) has(s INode) (ok bool) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	nodeID := s.Node().id
	for x := rh.minHeight; x <= rh.maxHeight; x++ {
		if rh.heights[x].has(nodeID) {
			ok = true
			return
		}
	}
	return
}

type recomputeHeapListIter struct {
	cursor INode
}

func (i *recomputeHeapListIter) Initialize(cursor INode) {
	i.cursor = cursor
}

func (i *recomputeHeapListIter) Next() (INode, bool) {
	if i.cursor == nil {
		return nil, false
	}
	prev := i.cursor
	pn := prev.Node()
	// the links carry *Node; hand back the owning interface value instead.
	if next := pn.nextInRecomputeHeap; next != nil {
		i.cursor = next.self
	} else {
		i.cursor = nil
	}
	pn.nextInRecomputeHeap = nil
	pn.previousInRecomputeHeap = nil
	pn.heightInRecomputeHeap = HeightUnset
	return prev, true
}

type RecomputeHeapListIterator interface {
	Initialize(INode)
	Next() (INode, bool)
}

func (rh *recomputeHeap) setIterToMinHeight(iter RecomputeHeapListIterator) {
	if iter == nil {
		return
	}

	rh.mu.Lock()
	defer rh.mu.Unlock()

	var heightBlock *recomputeHeapList
	for x := rh.minHeight; x < len(rh.heights); x++ {
		if rh.heights[x].count > 0 {
			heightBlock = &rh.heights[x]
			break
		}
	}
	if heightBlock == nil {
		return
	}
	iter.Initialize(heightBlock.head.self)
	heightBlock.head = nil
	heightBlock.tail = nil
	rh.numItems = rh.numItems - heightBlock.count
	heightBlock.count = 0
	rh.minHeight = rh.nextMinHeightUnsafe()
}

func (rh *recomputeHeap) remove(node INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.removeNodeUnsafe(node)
}

//
// utils
//

func (rh *recomputeHeap) removeMinUnsafe() (node INode, ok bool) {
	for x := rh.minHeight; x <= rh.maxHeight; x++ {
		block := &rh.heights[x]
		if block.count == 0 {
			continue
		}
		n := block.popNode()
		rh.numItems--
		n.heightInRecomputeHeap = HeightUnset
		if block.count > 0 {
			rh.minHeight = x
		} else {
			// the block just emptied, so the next candidate is above it; this
			// keeps the scan monotone rather than restarting from zero.
			rh.minHeight = rh.nextMinHeightFromUnsafe(x + 1)
		}
		return n.self, true
	}
	return
}

func (rh *recomputeHeap) addNodeUnsafe(s INode) {
	sn := s.Node()
	height := sn.height
	sn.heightInRecomputeHeap = height
	rh.maybeUpdateMinMaxHeightsUnsafe(height)
	rh.maybeAddNewHeightsUnsafe(height)
	sn.self = s
	rh.heights[height].pushNode(sn)
	rh.numItems++
}

func (rh *recomputeHeap) removeNodeUnsafe(item INode) {
	in := item.Node()
	rh.numItems--
	height := in.heightInRecomputeHeap
	block := &rh.heights[height]
	// the node is linked into the block for its own recorded height, so it can
	// be unlinked through its own pointers without searching the block.
	block.removeNode(in)
	if height == rh.minHeight && block.count == 0 {
		rh.minHeight = rh.nextMinHeightFromUnsafe(height + 1)
	}
	in.heightInRecomputeHeap = HeightUnset
}

func (rh *recomputeHeap) maybeUpdateMinMaxHeightsUnsafe(newHeight int) {
	if rh.numItems == 0 {
		rh.minHeight = newHeight
		rh.maxHeight = newHeight
		return
	}
	if rh.minHeight > newHeight {
		rh.minHeight = newHeight
	}
	if rh.maxHeight < newHeight {
		rh.maxHeight = newHeight
	}
}

func (rh *recomputeHeap) maybeAddNewHeightsUnsafe(newHeight int) {
	if len(rh.heights) <= newHeight {
		rh.heights = append(rh.heights, make([]recomputeHeapList, (newHeight-len(rh.heights))+1)...)
	}
}

func (rh *recomputeHeap) nextMinHeightUnsafe() (next int) {
	return rh.nextMinHeightFromUnsafe(0)
}

// minHeightUnsafe returns the lowest height holding a pending node, or [math.MaxInt]
// if the heap is empty.
//
// The tracked minHeight can lag behind the true minimum, since blocks are only
// rescanned as they drain. That is the safe direction for this function's
// callers: an answer lower than the truth makes them more conservative.
func (rh *recomputeHeap) minHeightUnsafe() int {
	if rh.numItems == 0 {
		return math.MaxInt
	}
	return rh.minHeight
}

// nextMinHeightFromUnsafe finds the lowest non-empty height at or above from.
//
// Callers that have just emptied a block pass the height above it: the minimum
// can only ever move up while a block is being drained, so resuming the scan
// there keeps draining the heap linear in the number of heights rather than
// rescanning every lower height on each removal.
func (rh *recomputeHeap) nextMinHeightFromUnsafe(from int) (next int) {
	if rh.numItems == 0 {
		return
	}
	if from < 0 {
		from = 0
	}
	for x := from; x < len(rh.heights); x++ {
		if rh.heights[x].count > 0 {
			return x
		}
	}
	return
}

func (rh *recomputeHeap) fixUnsafe(n INode) {
	rh.removeNodeUnsafe(n)
	rh.addNodeUnsafe(n)
}

// sanityCheck loops through each item in each height block
// and checks that all the height values match.
func (rh *recomputeHeap) sanityCheck() error {
	if rh.numItems > 0 && rh.heights[rh.minHeight].count == 0 {
		return fmt.Errorf("recompute heap; sanity check; lookup has items but min height block is empty")
	}
	for heightIndex := range rh.heights {
		cursor := rh.heights[heightIndex].head
		for cursor != nil {
			if cursor.heightInRecomputeHeap != heightIndex {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d", heightIndex, cursor.heightInRecomputeHeap)
			}
			if cursor.heightInRecomputeHeap != cursor.height {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d and node has height %d", heightIndex, cursor.heightInRecomputeHeap, cursor.height)
			}
			cursor = cursor.nextInRecomputeHeap
		}
	}
	return nil
}
