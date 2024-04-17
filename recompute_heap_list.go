package incr

// recomputeHeapList is a linked recomputeHeapList structure that can be used
// as a ordered recomputeHeapList as well as a constant time
// map using a similar technique to high throughput LRU queues.
type recomputeHeapList struct {
	head  INode
	tail  INode
	count int
}

func (l *recomputeHeapList) len() int {
	if l == nil {
		return 0
	}
	return l.count
}

func (l *recomputeHeapList) push(v INode) {
	l.count = l.count + 1
	v.Node().nextInRecomputeHeap = nil
	v.Node().previousInRecomputeHeap = nil
	if l.head == nil {
		l.head = v
		l.tail = v
		return
	}
	l.tail.Node().nextInRecomputeHeap = v
	v.Node().previousInRecomputeHeap = l.tail
	l.tail = v
}

func (l *recomputeHeapList) pop() (k Identifier, v INode, ok bool) {
	if l.head == nil {
		return
	}

	k = l.head.Node().id
	v = l.head
	ok = true
	l.count = l.count - 1

	if l.head == l.tail {
		l.head = nil
		l.tail = nil
		v.Node().nextInRecomputeHeap = nil
		v.Node().previousInRecomputeHeap = nil
		return
	}

	next := l.head.Node().nextInRecomputeHeap
	next.Node().previousInRecomputeHeap = nil
	l.head = next
	v.Node().nextInRecomputeHeap = nil
	v.Node().previousInRecomputeHeap = nil
	return
}

func (l *recomputeHeapList) consume(fn func(INode)) {
	cursor := l.head
	var next INode
	for cursor != nil {
		next = cursor.Node().nextInRecomputeHeap
		fn(cursor)
		cursor.Node().nextInRecomputeHeap = nil
		cursor.Node().previousInRecomputeHeap = nil
		cursor = next
	}
	l.head = nil
	l.tail = nil
	l.count = 0
}

func (l *recomputeHeapList) has(k Identifier) (ok bool) {
	if l == nil || l.head == nil {
		return
	}
	cursor := l.head
	for cursor != nil {
		if cursor.Node().id == k {
			ok = true
			return
		}
		cursor = cursor.Node().nextInRecomputeHeap
	}
	return
}

func (l *recomputeHeapList) find(k Identifier) (n INode, ok bool) {
	if l == nil || l.head == nil {
		return
	}
	cursor := l.head
	for cursor != nil {
		if cursor.Node().id == k {
			n = cursor
			ok = true
			return
		}
		cursor = cursor.Node().nextInRecomputeHeap
	}
	return
}

func (l *recomputeHeapList) remove(k Identifier) (ok bool) {
	if l.head == nil {
		return
	}

	var node INode
	node, ok = l.find(k)
	if !ok {
		return
	}
	l.count = l.count - 1
	if l.head == node {
		l.removeHeadItem()
	} else {
		l.removeLinkedItem(node)
	}
	return
}

func (l *recomputeHeapList) removeHeadItem() {
	if l.head == l.tail {
		l.head.Node().nextInRecomputeHeap = nil
		l.head.Node().previousInRecomputeHeap = nil
		l.head = nil
		l.tail = nil
		return
	}
	oldHead := l.head
	towardsTail := l.head.Node().nextInRecomputeHeap
	if towardsTail != nil {
		towardsTail.Node().previousInRecomputeHeap = nil
	}
	l.head = towardsTail
	oldHead.Node().nextInRecomputeHeap = nil
	oldHead.Node().previousInRecomputeHeap = nil
}

func (l *recomputeHeapList) removeLinkedItem(item INode) {
	towardsHead := item.Node().previousInRecomputeHeap
	towardsTail := item.Node().nextInRecomputeHeap
	if towardsHead != nil {
		towardsHead.Node().nextInRecomputeHeap = towardsTail
	}
	if towardsTail != nil {
		towardsTail.Node().previousInRecomputeHeap = towardsHead
	}
	if l.tail == item {
		l.tail = item.Node().previousInRecomputeHeap
		if l.tail != nil {
			l.tail.Node().nextInRecomputeHeap = nil
		}
	}
	item.Node().previousInRecomputeHeap = nil
	item.Node().nextInRecomputeHeap = nil
}
