package incr

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// newNodeHeap returns a new recompute heap with a given maximum height.
func newNodeHeap(initialHeights int) *nodeHeap {
	return &nodeHeap{
		heights: make([]map[Identifier]*recomputeHeapItem, initialHeights),
		lookup:  make(map[Identifier]*recomputeHeapItem),
	}
}

// nodeHeap is a height ordered list of lists of nodes.
type nodeHeap struct {
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

// Clear completely resets the recompute heap, preserving
// its current capacity.
func (rh *nodeHeap) Clear() {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.heights = make([]map[Identifier]*recomputeHeapItem, len(rh.heights))
	clear(rh.lookup)
	rh.minHeight = 0
	rh.maxHeight = 0
}

// MinHeight is the minimum height in the heap with nodes.
func (rh *nodeHeap) MinHeight() int {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	return rh.minHeight
}

// MinHeight is the minimum height in the heap with nodes.
func (rh *nodeHeap) MaxHeight() int {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	return rh.maxHeight
}

// Len returns the length of the recompute heap.
func (rh *nodeHeap) Len() int {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	return len(rh.lookup)
}

// Add adds nodes to the recompute heap.
func (rh *nodeHeap) Add(nodes ...INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	rh.addUnsafe(nodes...)
}

// Fix moves an existing node around in the height lists if its height has changed.
func (rh *nodeHeap) Fix(ids ...Identifier) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.fixUnsafe(ids...)
}

// Has returns if a given node exists in the recompute heap at its height by id.
func (rh *nodeHeap) Has(s INode) (ok bool) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	_, ok = rh.lookup[s.Node().id]
	return
}

// RemoveMinHeight removes the minimum height nodes from
// the recompute heap all at once.
func (rh *nodeHeap) RemoveMinHeight() (nodes []*recomputeHeapItem) {
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

func first[K comparable, V any](m map[K]V) (key K, out V, ok bool) {
	for key, out = range m {
		ok = true
		return
	}
	return
}

// RemoveMin removes a single minimum height node.
func (rh *nodeHeap) RemoveMin() (node *recomputeHeapItem) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	var id Identifier
	if rh.heights[rh.minHeight] != nil && len(rh.heights[rh.minHeight]) > 0 {
		id, node, _ = first(rh.heights[rh.minHeight])
		delete(rh.lookup, id)
		delete(rh.heights[rh.minHeight], id)
		if len(rh.heights[rh.minHeight]) == 0 {
			rh.minHeight = rh.nextMinHeightUnsafe()
		}
	}
	return
}

// Remove removes a specific node from the heap.
func (rh *nodeHeap) Remove(s INode) (ok bool) {
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

func (rh *nodeHeap) fixUnsafe(ids ...Identifier) {
	for _, id := range ids {
		if item, ok := rh.lookup[id]; ok {
			delete(rh.heights[item.height], id)
			rh.addNodeUnsafe(item.node)
		}
	}
}

func (rh *nodeHeap) addUnsafe(nodes ...INode) {
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

func (rh *nodeHeap) addNodeUnsafe(s INode) {
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

func (rh *nodeHeap) removeItemUnsafe(item *recomputeHeapItem) {
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

func (rh *nodeHeap) maybeUpdateMinMaxHeights(newHeight int) {
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

func (rh *nodeHeap) maybeAddNewHeights(newHeight int) {
	if len(rh.heights) <= newHeight {
		required := (newHeight - len(rh.heights)) + 1
		for x := 0; x < required; x++ {
			rh.heights = append(rh.heights, nil)
		}
	}
}

// nextMinHeightUnsafe finds the next smallest height in the heap that has nodes.
func (rh *nodeHeap) nextMinHeightUnsafe() (next int) {
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
func (rh *nodeHeap) sanityCheck() error {
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

func (rh *nodeHeap) String() string {
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
