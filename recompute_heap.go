package incr

import (
	"fmt"
	"sync"
)

// newRecomputeHeap returns a new recompute heap with a given maximum height.
func newRecomputeHeap(maxHeight int) *recomputeHeap {
	return &recomputeHeap{
		heights: make([]*recomputeHeapList, maxHeight),
	}
}

// recomputeHeap is a height ordered list of lists of nodes.
type recomputeHeap struct {
	// mu synchronizes critical sections for the heap.
	mu sync.Mutex

	// minHeight is the smallest heights index that has nodes
	minHeight int
	// maxHeight is the largest heights index that has nodes
	maxHeight int

	// heights is an array of linked lists corresponding
	// to node heights. it should be pre-allocated with
	// the constructor to the height limit number of elements.
	heights []*recomputeHeapList

	numItems int
}

// clear completely resets the recompute heap, preserving
// its current capacity.
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

	rh.addUnsafe(nodes...)
}

func (rh *recomputeHeap) fix(node INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	if rh.heights[node.Node().heightInRecomputeHeap].remove(node.Node().id) {
		rh.numItems--
	}
	rh.addNodeUnsafe(node)
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

func (rh *recomputeHeap) removeMinHeight() (nodes []INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	if rh.heights[rh.minHeight] != nil && rh.heights[rh.minHeight].len() > 0 {
		heightLen := rh.heights[rh.minHeight].len()
		nodes = make([]INode, 0, heightLen)
		rh.heights[rh.minHeight].consume(func(id Identifier, n INode) {
			if n.Node().height != n.Node().heightInRecomputeHeap {
				panic(fmt.Errorf("bad consume node height; %v vs. %v", n, n.Node().heightInRecomputeHeap))
			}
			n.Node().heightInRecomputeHeap = HeightUnset
			nodes = append(nodes, n)
		})
		if len(nodes) != heightLen {
			panic(fmt.Errorf("bad consume; %v vs. %v", len(nodes), heightLen))
		}
		rh.numItems -= len(nodes)
		rh.minHeight = rh.nextMinHeightUnsafe()
		return
	}
	return
}

func (rh *recomputeHeap) remove(node INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	rh.removeItemUnsafe(node)
	return
}

//
// utils
//

// removeMinUnsafe removes the minimum height node.
func (rh *recomputeHeap) removeMinUnsafe() (node INode, ok bool) {
	for x := rh.minHeight; x <= rh.maxHeight; x++ {
		if rh.heights[x] != nil && rh.heights[x].len() > 0 {
			_, node, ok = rh.heights[x].pop()
			node.Node().heightInRecomputeHeap = HeightUnset
			rh.numItems--
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

func (rh *recomputeHeap) addUnsafe(nodes ...INode) {
	for _, s := range nodes {
		rh.addNodeUnsafe(s)
	}
}

func (rh *recomputeHeap) addNodeUnsafe(s INode) {
	sn := s.Node()
	height := sn.height
	s.Node().heightInRecomputeHeap = height
	rh.maybeUpdateMinMaxHeights(height)
	rh.maybeAddNewHeightsUnsafe(height)
	if rh.heights[height] == nil {
		rh.heights[height] = new(recomputeHeapList)
	}
	rh.heights[height].push(s)
	rh.numItems++
}

func (rh *recomputeHeap) removeItemUnsafe(item INode) {
	id := item.Node().id
	height := item.Node().heightInRecomputeHeap
	rh.numItems--
	rh.heights[height].remove(id)
	isLastAtHeight := rh.heights[height].len() == 0
	if height == rh.minHeight && isLastAtHeight {
		rh.minHeight = rh.nextMinHeightUnsafe()
	}
	item.Node().heightInRecomputeHeap = HeightUnset
}

func (rh *recomputeHeap) maybeUpdateMinMaxHeights(newHeight int) {
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

// nextMinHeightUnsafe finds the next smallest height in the heap that has nodes.
func (rh *recomputeHeap) nextMinHeightUnsafe() (next int) {
	if rh.numItems == 0 {
		return
	}
	for x := 0; x <= rh.maxHeight; x++ {
		if rh.heights[x] != nil && rh.heights[x].len() > 0 {
			next = x
			break
		}
	}
	return
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
