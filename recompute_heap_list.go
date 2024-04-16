package incr

// recomputeHeapList is a linked recomputeHeapList structure that can be used
// as a ordered recomputeHeapList as well as a constant time
// map using a similar technique to high throughput LRU queues.
type recomputeHeapList struct {
	// head is the "first" element in the list
	head INode
	// tail is the "last" element in the list
	tail INode
}

// func (l *recomputeHeapList) isEmpty() bool {
// 	return len(l.items) == 0
// }

func (l *recomputeHeapList) isEmpty() bool {
	return l == nil || l.head == nil
}

func (l *recomputeHeapList) len() int {
	if l == nil {
		return 0
	}
	count := 0
	for v := l.head; v != nil; v = v.Node().nextInRecomputeHeap {
		count++
	}

	return count
}

func (l *recomputeHeapList) push(v INode) {
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

func (l *recomputeHeapList) consume(fn func(Identifier, INode)) {
	for v := l.head; v != nil; {
		next := v.Node().nextInRecomputeHeap
		v.Node().nextInRecomputeHeap = nil
		v.Node().previousInRecomputeHeap = nil
		fn(v.Node().ID(), v)
		v = next
	}

	l.head = nil
	l.tail = nil
}

func (l *recomputeHeapList) has(k Identifier) (ok bool) {
	if l == nil {
		return
	}
	for v := l.head; v != nil; v = v.Node().nextInRecomputeHeap {
		if v.Node().id == k {
			return true
		}
	}

	return
}

func (l *recomputeHeapList) remove(k Identifier) bool {
	var node INode
	for v := l.head; v != nil; v = v.Node().nextInRecomputeHeap {
		if v.Node().id == k {
			node = v
			break
		}
	}

	if node == nil {
		return false
	}

	if l.head == node {
		l.removeHeadItem()
	} else {
		l.removeLinkedItem(node)
	}
	return true
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
