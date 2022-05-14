package incr

import (
	"fmt"
	"strings"
	"sync"
)

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

// Add adds nodes to the recompute heap.
func (rh *recomputeHeap) Add(nodes ...INode) {
	rh.mu.Lock()
	defer rh.mu.Unlock()
	rh.addUnsafe(nodes...)
}

func (rh *recomputeHeap) addUnsafe(nodes ...INode) {
	for _, s := range nodes {
		sn := s.Node()
		if sn.height >= rh.heightLimit {
			panic("recompute heap; cannot add node with height greater than max height")
		}
		if _, ok := rh.lookup[sn.id]; ok {
			continue
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
	// head is the "first" element in the list
	head *recomputeHeapListItem
	// head is the "last" element in the list
	tail *recomputeHeapListItem
	// len is the overall length of the list
	len int
}

// String implements fmt.Stringer.
func (rhl *recomputeHeapList) String() string {
	var nodes []string
	ptr := rhl.head
	for ptr != nil {
		nodes = append(nodes, fmt.Sprint(ptr.key.Short()))
		ptr = ptr.next
	}
	return strings.Join(nodes, "->")
}

// push appends a node to the end, or tail, of the recompute heap list
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

	// rhl.tail here may be the head
	rhl.tail.next = item
	item.previous = rhl.tail
	item.next = nil
	rhl.tail = item
	return item
}

// pop removes the first, or head, of the recompute list
func (rhl *recomputeHeapList) pop() (k Identifier, v INode) {
	if rhl.head == nil {
		return
	}
	rhl.len--
	k = rhl.head.key
	v = rhl.head.value

	// specific case when we have (1) element left
	if rhl.head == rhl.tail {
		rhl.head = nil
		rhl.tail = nil
		return
	}

	// set the head to whatever the element after the head is
	next := rhl.head.next
	next.previous = nil
	rhl.head = next
	return
}

func (rhl *recomputeHeapList) popAll() (output []INode) {
	ptr := rhl.head
	for ptr != nil {
		output = append(output, ptr.value)
		ptr = ptr.next
	}
	rhl.head = nil
	rhl.tail = nil
	rhl.len = 0
	return
}

func (rhl *recomputeHeapList) find(n INode) *recomputeHeapListItem {
	nodeID := n.Node().id
	ptr := rhl.head
	for ptr != nil {
		if ptr.key == nodeID {
			return ptr
		}
		ptr = ptr.next
	}
	return nil
}

func (rhl *recomputeHeapList) remove(i *recomputeHeapListItem) {
	rhl.len--

	// three possibilities
	// - i is both the head and the tail
	// 		- nil out both
	// - i is the head
	// 		- set the head to i's next
	// - i is the tail
	//		- set the tail to i's previous
	// - i is neither
	//		- if i has a next, set its previous to i's previous
	//		- if i has a previous, set its previous to i's next

	if rhl.head == i && rhl.tail == i {
		rhl.head = nil
		rhl.tail = nil
		return
	}
	if rhl.head == i {
		rhl.head = i.next
		return
	}
	if rhl.tail == i {
		rhl.tail = i.previous
		return
	}

	next := i.next
	if next != nil {
		next.previous = i.previous
	}
	previous := i.previous
	if previous != nil {
		previous.next = i.next
	}
}

type recomputeHeapListItem struct {
	// key is the INode identifier
	key Identifier
	// value is the INode
	value INode
	// next is the pointer towards the tail
	next *recomputeHeapListItem
	// previous is the pointer towards the head
	previous *recomputeHeapListItem
}
