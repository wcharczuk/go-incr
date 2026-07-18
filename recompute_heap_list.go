package incr

// recomputeHeapList is an intrusive doubly-linked list of the nodes at a single
// height in the recompute heap.
//
// The links live in *Node fields on the nodes themselves, so pushing, popping
// and removing are all O(1) pointer writes with no allocation. Threading the
// list through *Node rather than INode matters for throughput: walking or
// relinking a block would otherwise cost an interface method call per field
// access, and these are the hottest operations in stabilization.
type recomputeHeapList struct {
	head  *Node
	tail  *Node
	count int
}

func (l *recomputeHeapList) len() int {
	if l == nil {
		return 0
	}
	return l.count
}

func (l *recomputeHeapList) push(v INode) {
	n := v.Node()
	// the heap has to hand the owning interface value back to the graph when
	// this node is popped, and the links only carry *Node, so record it here.
	n.self = v
	l.pushNode(n)
}

func (l *recomputeHeapList) pushNode(n *Node) {
	l.count++
	n.nextInRecomputeHeap = nil
	n.previousInRecomputeHeap = l.tail
	if l.head == nil {
		l.head = n
		l.tail = n
		return
	}
	l.tail.nextInRecomputeHeap = n
	l.tail = n
}

func (l *recomputeHeapList) pop() (k Identifier, v INode, ok bool) {
	n := l.popNode()
	if n == nil {
		return
	}
	return n.id, n.self, true
}

func (l *recomputeHeapList) popNode() *Node {
	n := l.head
	if n == nil {
		return nil
	}
	l.count--
	next := n.nextInRecomputeHeap
	l.head = next
	if next == nil {
		l.tail = nil
	} else {
		next.previousInRecomputeHeap = nil
	}
	n.nextInRecomputeHeap = nil
	n.previousInRecomputeHeap = nil
	return n
}

func (l *recomputeHeapList) consume(fn func(INode)) {
	cursor := l.head
	for cursor != nil {
		next := cursor.nextInRecomputeHeap
		self := cursor.self
		cursor.nextInRecomputeHeap = nil
		cursor.previousInRecomputeHeap = nil
		fn(self)
		cursor = next
	}
	l.head = nil
	l.tail = nil
	l.count = 0
}

func (l *recomputeHeapList) has(k Identifier) (ok bool) {
	_, ok = l.find(k)
	return
}

func (l *recomputeHeapList) find(k Identifier) (n INode, ok bool) {
	if l == nil {
		return
	}
	cursor := l.head
	for cursor != nil {
		if cursor.id == k {
			return cursor.self, true
		}
		cursor = cursor.nextInRecomputeHeap
	}
	return
}

// removeNode unlinks a node that is known to be in this list.
//
// Callers reach the right list via the node's heightInRecomputeHeap, so
// membership is already established and no search is needed.
func (l *recomputeHeapList) removeNode(n *Node) {
	prev := n.previousInRecomputeHeap
	next := n.nextInRecomputeHeap
	if prev != nil {
		prev.nextInRecomputeHeap = next
	} else {
		l.head = next
	}
	if next != nil {
		next.previousInRecomputeHeap = prev
	} else {
		l.tail = prev
	}
	n.previousInRecomputeHeap = nil
	n.nextInRecomputeHeap = nil
	l.count--
}

// remove unlinks the node with a given id, searching the list to find it.
//
// Prefer removeNode where the node is already in hand; this exists for callers
// that only hold an identifier.
func (l *recomputeHeapList) remove(k Identifier) (ok bool) {
	var node INode
	node, ok = l.find(k)
	if !ok {
		return
	}
	l.removeNode(node.Node())
	return
}
