package incr

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// newRecomputeHeap returns a new recompute heap with a given maximum height.
func newRecomputeHeap(initialHeights int) *recomputeHeap {
	return &recomputeHeap{
		heights: make([]map[Identifier]*recomputeHeapItem, initialHeights),
		lookup:  make(map[Identifier]*recomputeHeapItem),
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
	heights []map[Identifier]*recomputeHeapItem
	// lookup is a quick lookup function for testing if an item exists
	// in the heap, and specifically removing single elements quickly by id.
	lookup map[Identifier]*recomputeHeapItem
}

type recomputeHeapItem struct {
	node   INode
	height int
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

// Fix moves an existing node around in the height lists if its height has changed.
func (rh *recomputeHeap) Fix(ids ...Identifier) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.fixUnsafe(ids...)
}

// Has returns if a given node exists in the recompute heap at its height by id.
func (rh *recomputeHeap) Has(s INode) (ok bool) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	_, ok = rh.lookup[s.Node().id]
	return
}

// RemoveMinHeight removes the minimum height nodes from
// the recompute heap all at once.
func (rh *recomputeHeap) RemoveMinHeight() (nodes []*recomputeHeapItem) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	if rh.heights[rh.minHeight] != nil && len(rh.heights[rh.minHeight]) > 0 {
		nodes = make([]*recomputeHeapItem, 0, len(rh.heights[rh.minHeight]))
		for id, n := range rh.heights[rh.minHeight] {
			nodes = append(nodes, n)
			delete(rh.lookup, id)
		}
		clear(rh.heights[rh.minHeight])
		rh.minHeight = rh.nextMinHeightUnsafe()
	}
	return
}

// Remove removes a specific node from the heap.
func (rh *recomputeHeap) Remove(s INode) (ok bool) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	sn := s.Node()
	var item *recomputeHeapItem
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

func (rh *recomputeHeap) fixUnsafe(ids ...Identifier) {
	for _, id := range ids {
		if item, ok := rh.lookup[id]; ok {
			delete(rh.heights[item.height], id)
			rh.addNodeUnsafe(item.node)
		}
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
		rh.heights[sn.height] = make(map[Identifier]*recomputeHeapItem)
	}
	item := &recomputeHeapItem{node: s, height: sn.height}
	rh.heights[sn.height][sn.id] = item
	rh.lookup[sn.id] = item
}

func (rh *recomputeHeap) removeItemUnsafe(item *recomputeHeapItem) {
	id := item.node.Node().id
	delete(rh.lookup, id)
	delete(rh.heights[item.height], id)

	// handle the edge case where removing a node removes the _last_ node
	// in the current minimum height list, causing us to need to move
	// the minimum height up one value.
	isLastAtHeight := rh.heights[item.height] == nil || len(rh.heights[item.height]) == 0
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
		if len(rh.heights[x]) > 0 {
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
		for _, item := range height {
			if item.height != heightIndex {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d", heightIndex, item.height)
			}
			if item.height != item.node.Node().height {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d and node has height %d", heightIndex, item.height, item.node.Node().height)
			}
		}
	}
	return nil
}

func (rh *recomputeHeap) String() string {
	output := new(bytes.Buffer)

	fmt.Fprintf(output, "{\n")
	for heightIndex, heightList := range rh.heights {
		if heightList == nil {
			// fmt.Fprintf(output, "\t%d: []\n", heightIndex)
			continue
		}
		fmt.Fprintf(output, "\t%d: [", heightIndex)
		lineParts := make([]string, 0, len(heightList))
		for _, li := range heightList {
			lineParts = append(lineParts, fmt.Sprint(li.node))
		}
		fmt.Fprintf(output, "%s],\n", strings.Join(lineParts, ", "))
	}
	fmt.Fprintf(output, "}\n")
	return output.String()
}
