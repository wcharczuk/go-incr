package incr

import (
	"fmt"
	"sync"
)

func newRecomputeHeap(maxHeight int) *recomputeHeap {
	return &recomputeHeap{
		heights: make([]*recomputeHeapList, maxHeight),
	}
}

type recomputeHeap struct {
	mu        sync.Mutex
	minHeight int
	maxHeight int
	heights   []*recomputeHeapList
	numItems  int
}

func (rh *recomputeHeap) clear() {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	rh.heights = make([]*recomputeHeapList, len(rh.heights))
	rh.minHeight = 0
	rh.maxHeight = 0
	rh.numItems = 0
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

func (i *recomputeHeapListIter) Next() (INode, bool) {
	if i.cursor == nil {
		return nil, false
	}
	prev := i.cursor
	i.cursor = i.cursor.Node().nextInRecomputeHeap
	prev.Node().nextInRecomputeHeap = nil
	prev.Node().previousInRecomputeHeap = nil
	prev.Node().heightInRecomputeHeap = HeightUnset
	return prev, true
}

func (rh *recomputeHeap) setIterToMinHeight(iter *recomputeHeapListIter) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	var heightBlock *recomputeHeapList
	for x := 0; x < len(rh.heights); x++ {
		heightBlock = rh.heights[x]
		if heightBlock != nil && heightBlock.len() > 0 {
			break
		}
	}
	iter.cursor = heightBlock.head
	heightBlock.head = nil
	heightBlock.tail = nil
	rh.numItems = rh.numItems - heightBlock.len()
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
		if rh.heights[x] != nil && rh.heights[x].len() > 0 {
			_, node, ok = rh.heights[x].pop()
			rh.numItems--
			node.Node().heightInRecomputeHeap = HeightUnset
			if rh.heights[x].len() > 0 {
				rh.minHeight = x
			} else {
				rh.minHeight = rh.nextMinHeightUnsafe()
			}
			return
		}
	}
	return
}

func (rh *recomputeHeap) addNodeUnsafe(s INode) {
	sn := s.Node()
	height := sn.height
	s.Node().heightInRecomputeHeap = height
	rh.maybeUpdateMinMaxHeightsUnsafe(height)
	rh.maybeAddNewHeightsUnsafe(height)
	if rh.heights[height] == nil {
		rh.heights[height] = new(recomputeHeapList)
	}
	rh.heights[height].push(s)
	rh.numItems++
}

func (rh *recomputeHeap) removeNodeUnsafe(item INode) {
	rh.numItems--
	id := item.Node().id
	height := item.Node().heightInRecomputeHeap
	rh.heights[height].remove(id)
	isLastAtHeight := rh.heights[height].len() == 0
	if height == rh.minHeight && isLastAtHeight {
		rh.minHeight = rh.nextMinHeightUnsafe()
	}
	item.Node().heightInRecomputeHeap = HeightUnset
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
		required := (newHeight - len(rh.heights)) + 1
		for x := 0; x < required; x++ {
			rh.heights = append(rh.heights, nil)
		}
	}
}

func (rh *recomputeHeap) nextMinHeightUnsafe() (next int) {
	if rh.numItems == 0 {
		return
	}
	for x := 0; x < len(rh.heights); x++ {
		if rh.heights[x] != nil && rh.heights[x].len() > 0 {
			next = x
			break
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
	if rh.numItems > 0 && (rh.heights[rh.minHeight] == nil || rh.heights[rh.minHeight].len() == 0) {
		return fmt.Errorf("recompute heap; sanity check; lookup has items but min height block is empty")
	}
	for heightIndex, height := range rh.heights {
		if height == nil {
			continue
		}
		cursor := height.head
		for cursor != nil {
			if cursor.Node().heightInRecomputeHeap != heightIndex {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d", heightIndex, cursor.Node().heightInRecomputeHeap)
			}
			if cursor.Node().heightInRecomputeHeap != cursor.Node().height {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d and node has height %d", heightIndex, cursor.Node().heightInRecomputeHeap, cursor.Node().height)
			}
			cursor = cursor.Node().nextInRecomputeHeap
		}
	}
	return nil
}
