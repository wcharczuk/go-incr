package incr

import (
	"fmt"
	"sync"
)

// defaultRecomputeHeapMaxHeight is the default
// maximum recompute heap height when we create graph states.
const defaultRecomputeHeapMaxHeight = 255

// newRecomputeHeap returns a new recompute heap with a given maximum height.
func newRecomputeHeap(heightLimit int) *recomputeHeap {
	return &recomputeHeap{
		heightLimit: heightLimit,
		heights:     make([]*list[Identifier, INode], heightLimit),
		lookup:      make(map[Identifier]*listItem[Identifier, INode]),
	}
}

// recomputeHeap is a height ordered list of lists of nodes.
type recomputeHeap struct {
	// mu synchronizes critical sections for the heap.
	mu sync.Mutex

	// heightLimit is the maximum height a node can be
	// in the recompute heap. it should also be
	// the length of the heights array.
	heightLimit int

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

	sn := s.Node()
	if sn.height >= rh.heightLimit {
		panic("recompute heap; cannot has node with height greater than max height")
	}
	_, ok = rh.lookup[sn.id]
	return
}

// RemoveMin removes the minimum node from the recompute heap.
func (rh *recomputeHeap) RemoveMin() INode {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	if rh.heights[rh.minHeight] != nil && rh.heights[rh.minHeight].Len() > 0 {
		id, node, _ := rh.heights[rh.minHeight].Pop()
		delete(rh.lookup, id)
		if rh.heights[rh.minHeight].Len() == 0 {
			rh.minHeight = rh.nextMinHeight()
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

	if rh.heights[rh.minHeight] != nil && rh.heights[rh.minHeight].Len() > 0 {
		nodes = rh.heights[rh.minHeight].PopAll()
		for _, n := range nodes {
			delete(rh.lookup, n.Node().id)
		}
		rh.minHeight = rh.nextMinHeight()
	}
	return
}

// Remove removes a specific node from the heap.
func (rh *recomputeHeap) Remove(s INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	sn := s.Node()
	item, ok := rh.lookup[sn.id]
	if !ok {
		return
	}
	delete(rh.lookup, sn.id)
	rh.heights[sn.height].Remove(item.key)

	if sn.height == rh.minHeight && (rh.heights[sn.height] == nil || rh.heights[sn.height].Len() == 0) {
		rh.minHeight = rh.nextMinHeight()
	}
}

//
// utils
//

func (rh *recomputeHeap) addUnsafe(nodes ...INode) {
	for _, s := range nodes {
		sn := s.Node()
		if sn.height >= rh.heightLimit {
			panic(fmt.Sprintf("recompute heap; cannot add node with height %d which is greater than the max height %d", sn.height, rh.heightLimit))
		}

		// this needs to be here for `SetStale` to work
		// correctly, specifically we may need to
		// add nodes to the recompute heap multiple times.
		if _, ok := rh.lookup[sn.id]; ok {
			// we may need to move the node
			if rh.heights[sn.height] == nil || !rh.heights[sn.height].Has(sn.id) {
				// remove it from any existing height lists
				for _, height := range rh.heights {
					if height != nil {
						height.Remove(sn.id)
					}
				}
				// add it to the correct one
				rh.addNodeUnsafe(s)
			}
			continue
		}
		rh.addNodeUnsafe(s)
	}
}

func (rh *recomputeHeap) addNodeUnsafe(s INode) {
	sn := s.Node()
	// when we add nodes, make sure to note if we
	// need to change the min or max height
	if len(rh.lookup) == 0 {
		rh.minHeight = sn.height
		rh.maxHeight = sn.height
	} else if rh.minHeight > sn.height {
		rh.minHeight = sn.height
	} else if rh.maxHeight < sn.height {
		rh.maxHeight = sn.height
	}
	if rh.heights[sn.height] == nil {
		rh.heights[sn.height] = new(list[Identifier, INode])
	}
	item := rh.heights[sn.height].Push(s.Node().id, s)
	rh.lookup[sn.id] = item
}

// nextMinHeight finds the next smallest height in the heap that has nodes.
func (rh *recomputeHeap) nextMinHeight() (next int) {
	if len(rh.lookup) == 0 {
		return
	}
	for x := rh.minHeight; x <= rh.maxHeight; x++ {
		if rh.heights[x] != nil && rh.heights[x].head != nil {
			next = x
			return
		}
	}
	return
}
