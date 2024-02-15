package incr

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// newRecomputeHeap returns a new recompute heap with a given maximum height.
func newRecomputeHeap(maxHeight int) *recomputeHeap {
	return &recomputeHeap{
		heights: make([]*list[Identifier, INode], maxHeight),
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

	numItems int
}

// clear completely resets the recompute heap, preserving
// its current capacity.
func (rh *recomputeHeap) clear() {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.heights = make([]*list[Identifier, INode], len(rh.heights))
	rh.minHeight = 0
	rh.maxHeight = 0
	rh.numItems = 0
}

func (rh *recomputeHeap) len() int {
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
	rh.fixUnsafe(node)
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

// removeMin removes the minimum height node.
func (rh *recomputeHeap) removeMin() (node INode, ok bool) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	node, ok = rh.removeMinUnsafe()
	return
}

// removeMinHeight removes the minimum height nodes from
// the recompute heap all at once.
func (rh *recomputeHeap) removeMinHeight() (nodes []INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	if rh.heights[rh.minHeight] != nil && rh.heights[rh.minHeight].len() > 0 {
		nodes = make([]INode, 0, rh.heights[rh.minHeight].len())
		rh.heights[rh.minHeight].consume(func(id Identifier, n INode) {
			n.Node().heightInRecomputeHeap = heightUnset
			nodes = append(nodes, n)
			rh.numItems--
		})
		rh.minHeight = rh.nextMinHeightUnsafe()
	}
	return
}

func (rh *recomputeHeap) remove(node INode) (ok bool) {
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
			node.Node().heightInRecomputeHeap = heightUnset
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

func (rh *recomputeHeap) fixUnsafe(node INode) {
	rh.heights[node.Node().heightInRecomputeHeap].remove(node.Node().id)
	rh.numItems--
	rh.addNodeUnsafe(node)
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
		rh.heights[height] = new(list[Identifier, INode])
	}
	rh.heights[height].push(sn.id, s)
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
	item.Node().heightInRecomputeHeap = heightUnset
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
	for x := rh.minHeight; x <= rh.maxHeight; x++ {
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
			item := cursor.value
			if item.Node().heightInRecomputeHeap != heightIndex {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d", heightIndex, item.Node().heightInRecomputeHeap)
			}
			if item.Node().heightInRecomputeHeap != item.Node().height {
				return fmt.Errorf("recompute heap; sanity check; at height %d item has height %d and node has height %d", heightIndex, item.Node().heightInRecomputeHeap, item.Node().height)
			}
			cursor = cursor.next
		}
	}
	return nil
}

func (rh *recomputeHeap) String() string {
	output := new(bytes.Buffer)

	fmt.Fprintf(output, "{\n")
	for heightIndex, heightList := range rh.heights {
		if heightList == nil {
			continue
		}
		fmt.Fprintf(output, "\t%d: [", heightIndex)
		lineParts := make([]string, 0, heightList.len())
		cursor := heightList.head
		for cursor != nil {
			lineParts = append(lineParts, fmt.Sprint(cursor.value))
			cursor = cursor.next
		}
		fmt.Fprintf(output, "%s],\n", strings.Join(lineParts, ", "))
	}
	fmt.Fprintf(output, "}\n")
	return output.String()
}
