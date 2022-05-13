package incr

import "sync"

// newRecomputeHeap returns a new recompute heap with a given maximum height.
func newRecomputeHeap(heightLimit int) *recomputeHeap {
	return &recomputeHeap{
		heightLimit: heightLimit,
		heights:     make([]*recomputeHeapList, heightLimit),
		lookup:      make(map[Identifier]*recomputeHeapListItem),
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
	// minHeight is the largest heights index that has nodes
	maxHeight int

	// heights is an array of linked lists corresponding
	// to node heights. it should be pre-allocated with
	// the constructor to the height limit number of elements.
	heights []*recomputeHeapList
	// lookup is a quick lookup function for testing if an item exists
	// in the heap, and specifically removing single elements quickly by id.
	lookup map[Identifier]*recomputeHeapListItem
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

// Add adds a node to the recompute heap at a given height.
func (rh *recomputeHeap) Add(s INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()

	sn := s.Node()
	if sn.height >= rh.heightLimit {
		panic("recompute heap; cannot add node with height greater than max height")
	}

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
		rh.heights[sn.height] = new(recomputeHeapList)
	}
	item := rh.heights[sn.height].push(s)
	rh.lookup[sn.id] = item
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

	if rh.heights[rh.minHeight] != nil && rh.heights[rh.minHeight].head != nil {
		id, node := rh.heights[rh.minHeight].pop()
		delete(rh.lookup, id)
		if rh.heights[rh.minHeight].head == nil {
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

	if rh.heights[rh.minHeight] != nil && rh.heights[rh.minHeight].head != nil {
		nodes = rh.heights[rh.minHeight].popAll()
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
	rh.heights[sn.height].remove(item)

	if sn.height == rh.minHeight && rh.heights[sn.height].head == nil {
		rh.minHeight = rh.nextMinHeight()
	}
}

// nextMinHeight finds the next smallest height in the heap that has nodes.
func (rh *recomputeHeap) nextMinHeight() (next int) {
	if len(rh.lookup) == 0 {
		return
	}
	for x := 0; x <= rh.maxHeight; x++ {
		if rh.heights[x] != nil && rh.heights[x].head != nil {
			next = x
			return
		}
	}
	return
}

type recomputeHeapList struct {
	head *recomputeHeapListItem
	tail *recomputeHeapListItem
	len  int
}

func (rhl *recomputeHeapList) push(v INode) *recomputeHeapListItem {
	item := &recomputeHeapListItem{
		key:   v.Node().id,
		value: v,
	}
	rhl.len++
	if rhl.head == nil {
		rhl.head = item
		rhl.tail = item
		return item
	}
	rhl.tail.previous = item
	item.next = rhl.tail
	rhl.tail = item
	return item
}

func (rhl *recomputeHeapList) pop() (k Identifier, v INode) {
	if rhl.head == nil {
		return
	}
	rhl.len--
	k = rhl.head.key
	v = rhl.head.value

	if rhl.head == rhl.tail {
		rhl.head = nil
		rhl.tail = nil
		return
	}
	after := rhl.head.previous
	if after != nil {
		after.next = nil
	}
	rhl.head = after
	return
}

func (rhl *recomputeHeapList) popAll() (output []INode) {
	for rhl.head != nil {
		output = append(output, rhl.head.value)
		rhl.head = rhl.head.previous
		rhl.len--
	}
	if rhl.head == nil {
		rhl.tail = nil
	}
	return
}

func (rhl *recomputeHeapList) remove(i *recomputeHeapListItem) {
	rhl.len--
	after := i.previous
	before := i.next
	if after != nil {
		after.next = before
	}
	if before != nil {
		before.previous = after
	}
	if rhl.tail == i {
		rhl.tail = i.next
		if rhl.tail != nil {
			rhl.tail.previous = nil
		}
	}
}

type recomputeHeapListItem struct {
	key   Identifier
	value INode

	next     *recomputeHeapListItem
	previous *recomputeHeapListItem
}
