package incr

// newRecomputeHeap returns a new recompute heap with a given maximum height.
func newRecomputeHeap(maxHeight int) *recomputeHeap {
	return &recomputeHeap{
		maxHeight: maxHeight,
		heights:   make([]*recomputeHeapList, maxHeight),
		lookup:    make(map[Identifier]*recomputeHeapListItem),
	}
}

// recomputeHeap is a height ordered list of lists of nodes.
//
// NOTE(wc): we should really switch this to a structure more similar
// to a lru cache, where we keep a linked list of nodes per "height"
// and a lookup map, that links to the node in its place in a height list.
type recomputeHeap struct {
	maxHeight int

	heights []*recomputeHeapList
	lookup  map[Identifier]*recomputeHeapListItem
}

type recomputeHeapList struct {
	head *recomputeHeapListItem
	tail *recomputeHeapListItem
	len  int
}

func (rhl *recomputeHeapList) push(v GraphNode) *recomputeHeapListItem {
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

func (rhl *recomputeHeapList) pop() (k Identifier, v GraphNode) {
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
	value GraphNode

	next     *recomputeHeapListItem
	previous *recomputeHeapListItem
}

func (rh *recomputeHeap) len() int { return len(rh.lookup) }

// add adds a node to the recompute heap at a given height.
func (rh *recomputeHeap) add(s GraphNode) {
	sn := s.Node()
	if sn.height >= rh.maxHeight {
		panic("recompute heap; cannot add node with height greater than max height")
	}
	if rh.heights[sn.height] == nil {
		rh.heights[sn.height] = new(recomputeHeapList)
	}
	item := rh.heights[sn.height].push(s)
	rh.lookup[sn.id] = item
}

// has returns if a given node exists in the recompute heap at its height by id
//
// there is an opportunity here to optimize this better with a lookup.
func (rh *recomputeHeap) has(s GraphNode) (ok bool) {
	sn := s.Node()
	if sn.height >= rh.maxHeight {
		panic("recompute heap; cannot has node with height greater than max height")
	}
	_, ok = rh.lookup[sn.id]
	return
}

// removeMin removes the minimum node from the recompute heap.
func (rh *recomputeHeap) removeMin() GraphNode {
	for height := range rh.heights {
		if rh.heights[height] != nil && rh.heights[height].head != nil {
			id, node := rh.heights[height].pop()
			delete(rh.lookup, id)
			return node
		}
	}
	return nil
}

// remove removes a specific node from the heap.
func (rh *recomputeHeap) remove(s GraphNode) {
	sn := s.Node()

	item, ok := rh.lookup[sn.id]
	if !ok {
		return
	}
	delete(rh.lookup, sn.id)
	rh.heights[sn.height].remove(item)
}
