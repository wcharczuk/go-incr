package incr

// adjustHeightsHeapList is a linked adjustHeightsHeapList structure that can be used
// as a ordered adjustHeightsHeapList as well as a constant time
// map using a similar technique to high throughput LRU queues.
type adjustHeightsHeapList struct {
	// head is the "first" element in the list
	head INode
	// tail is the "last" element in the list
	tail INode
	// items is a map between the key and the actual list item(s)
	items map[Identifier]INode
}

func (l *adjustHeightsHeapList) isEmpty() bool {
	return len(l.items) == 0
}

func (l *adjustHeightsHeapList) len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

func (l *adjustHeightsHeapList) push(v INode) {
	if l.items == nil {
		l.items = make(map[Identifier]INode)
	}
	l.items[v.Node().id] = v
	if l.head == nil {
		l.head = v
		l.tail = v
		return
	}
	l.tail.Node().nextInAdjustHeightsHeap = v
	v.Node().previousInAdjustHeightsHeap = l.tail
	l.tail = v
	return
}

func (l *adjustHeightsHeapList) pop() (k Identifier, v INode, ok bool) {
	if l.head == nil { // follows that items empty
		return
	}

	k = l.head.Node().id
	v = l.head
	ok = true
	delete(l.items, k)

	if l.head == l.tail {
		l.head = nil
		l.tail = nil
		v.Node().nextInAdjustHeightsHeap = nil
		v.Node().previousInAdjustHeightsHeap = nil
		return
	}

	next := l.head.Node().nextInAdjustHeightsHeap
	next.Node().previousInAdjustHeightsHeap = nil
	l.head = next

	v.Node().nextInAdjustHeightsHeap = nil
	v.Node().previousInAdjustHeightsHeap = nil
	return
}

func (l *adjustHeightsHeapList) has(k Identifier) (ok bool) {
	if l == nil || l.items == nil {
		return
	}
	_, ok = l.items[k]
	return
}

func (l *adjustHeightsHeapList) remove(k Identifier) (ok bool) {
	if len(l.items) == 0 {
		return
	}

	var node INode
	node, ok = l.items[k]
	if !ok {
		return
	}

	delete(l.items, k)
	if l.head == node {
		l.removeHeadItem()
	} else {
		l.removeLinkedItem(node)
	}
	return
}

func (l *adjustHeightsHeapList) removeHeadItem() {
	if l.head == l.tail {
		l.head.Node().nextInAdjustHeightsHeap = nil
		l.head.Node().previousInAdjustHeightsHeap = nil
		l.head = nil
		l.tail = nil
		return
	}
	oldHead := l.head
	towardsTail := l.head.Node().nextInAdjustHeightsHeap
	if towardsTail != nil {
		towardsTail.Node().previousInAdjustHeightsHeap = nil
	}
	l.head = towardsTail
	oldHead.Node().nextInAdjustHeightsHeap = nil
	oldHead.Node().previousInAdjustHeightsHeap = nil
}

func (l *adjustHeightsHeapList) removeLinkedItem(item INode) {
	towardsHead := item.Node().previousInAdjustHeightsHeap
	towardsTail := item.Node().nextInAdjustHeightsHeap
	if towardsHead != nil {
		towardsHead.Node().nextInAdjustHeightsHeap = towardsTail
	}
	if towardsTail != nil {
		towardsTail.Node().previousInAdjustHeightsHeap = towardsHead
	}
	if l.tail == item {
		l.tail = item.Node().previousInAdjustHeightsHeap
		if l.tail != nil {
			l.tail.Node().nextInAdjustHeightsHeap = nil
		}
	}
	item.Node().previousInAdjustHeightsHeap = nil
	item.Node().nextInAdjustHeightsHeap = nil
}
