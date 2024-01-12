package incr

import (
	"fmt"
	"sync"
)

// newRecomputeHeap returns a new recompute heap with a given maximum height.
func newRecomputeHeap(initialHeights int) *recomputeHeap {
	return &recomputeHeap{
		heights: make([]*list[Identifier, INode], initialHeights),
		lookup:  make(map[Identifier]*listItem[Identifier, INode]),
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
	heights []*list[Identifier, INode]
	// lookup is a quick lookup function for testing if an item exists
	// in the heap, and specifically removing single elements quickly by id.
	lookup map[Identifier]*listItem[Identifier, INode]
}

// MinHeight is the minimum height in the heap with nodes.
func (rh *recomputeHeap) MinHeight() int {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	return rh.minHeight
}

// MinHeight is the minimum height in the heap with nodes.
func (rh *recomputeHeap) MaxHeight() int {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	return rh.maxHeight
}

// Len returns the length of the recompute heap.
func (rh *recomputeHeap) Len() int {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	return len(rh.lookup)
}

// Add adds nodes to the recompute heap.
func (rh *recomputeHeap) Add(nodes ...INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.addUnsafe(nodes...)
}

// Has returns if a given node exists in the recompute heap at its height by id.
func (rh *recomputeHeap) Has(s INode) (ok bool) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	_, ok = rh.lookup[s.Node().id]
	return
}

// RemoveMin removes the minimum node from the recompute heap.
func (rh *recomputeHeap) RemoveMin() INode {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	if rh.heights[rh.minHeight] != nil && rh.heights[rh.minHeight].lenUnsafe() > 0 {
		id, node, _ := rh.heights[rh.minHeight].popUnsafe()
		delete(rh.lookup, id)
		if rh.heights[rh.minHeight].lenUnsafe() == 0 {
			rh.minHeight = rh.nextMinHeightUnsafe()
		}
		return node
	}
	return nil
}

// RemoveMinHeight removes the minimum height nodes from
// the recompute heap all at once.
func (rh *recomputeHeap) RemoveMinHeight() (nodes []INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	if rh.heights[rh.minHeight] != nil && rh.heights[rh.minHeight].lenUnsafe() > 0 {
		nodes = rh.heights[rh.minHeight].popAllUnsafe()
		for _, n := range nodes {
			delete(rh.lookup, n.Node().id)
		}
		rh.minHeight = rh.nextMinHeightUnsafe()
	}
	return
}

// Remove removes a specific node from the heap.
func (rh *recomputeHeap) Remove(s INode) (ok bool) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	sn := s.Node()
	var item *listItem[Identifier, INode]
	item, ok = rh.lookup[sn.id]
	if !ok {
		return
	}
	rh.removeItemUnsafe(item)
	return
}

//
// utils
//

// hasKey returns if a given node id exists in the recompute heap lookup.
func (rh *recomputeHeap) hasKey(id Identifier) (ok bool) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	_, ok = rh.lookup[id]
	return
}

// fix moves an existing node around in the height maps if its height changes.
func (rh *recomputeHeap) fix(id Identifier) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	if item, ok := rh.lookup[id]; ok {
		_ = rh.heights[item.height].removeUnsafe(item.key)
		rh.addNodeUnsafe(item.value)
	}
}

func (rh *recomputeHeap) addUnsafe(nodes ...INode) {
	for _, s := range nodes {
		sn := s.Node()
		// this needs to be here for `SetStale` to work correctly, specifically
		// we may need to add nodes to the recompute heap multiple times before
		// we ultimately call stabilize, and the heights may change during that time.
		if current, ok := rh.lookup[sn.id]; ok {
			rh.removeItemUnsafe(current)
		}
		rh.addNodeUnsafe(s)
	}
}

func (rh *recomputeHeap) addNodeUnsafe(s INode) {
	sn := s.Node()
	rh.maybeUpdateMinMaxHeights(sn.height)
	rh.maybeAddNewHeights(sn.height)
	if rh.heights[sn.height] == nil {
		rh.heights[sn.height] = new(list[Identifier, INode])
	}
	item := rh.heights[sn.height].pushUnsafe(sn.id, s, sn.height)
	rh.lookup[sn.id] = item
}

func (rh *recomputeHeap) removeItemUnsafe(item *listItem[Identifier, INode]) {
	delete(rh.lookup, item.key)
	rh.heights[item.height].removeUnsafe(item.key)

	// handle the edge case where removing a node removes the _last_ node
	// in the current minimum height, causing us to need to move
	// the minimum height up one value.
	isLastAtHeight := rh.heights[item.height] == nil || rh.heights[item.height].Len() == 0
	if item.height == rh.minHeight && isLastAtHeight {
		rh.minHeight = rh.nextMinHeightUnsafe()
	}
}

func (rh *recomputeHeap) maybeUpdateMinMaxHeights(newHeight int) {
	if len(rh.lookup) == 0 {
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

func (rh *recomputeHeap) maybeAddNewHeights(newHeight int) {
	if len(rh.heights) <= newHeight {
		required := (newHeight - len(rh.heights)) + 1
		for x := 0; x < required; x++ {
			rh.heights = append(rh.heights, nil)
		}
	}
}

// nextMinHeightUnsafe finds the next smallest height in the heap that has nodes.
func (rh *recomputeHeap) nextMinHeightUnsafe() (next int) {
	if len(rh.lookup) == 0 {
		return
	}
	for x := rh.minHeight; x <= rh.maxHeight; x++ {
		if rh.heights[x] != nil && rh.heights[x].head != nil {
			next = x
			break
		}
	}
	return
}

// sanityCheck loops through each item in each height block
// and checks that all the height values match.
func (rh *recomputeHeap) sanityCheck() error {
	for heightIndex, height := range rh.heights {
		if height == nil {
			continue
		}
		for _, item := range height.items {
			if item.height != heightIndex {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d", heightIndex, item.height)
			}
			if item.height != item.value.Node().height {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d and node has height %d", heightIndex, item.height, item.value.Node().height)
			}
		}
	}
	return nil
}
