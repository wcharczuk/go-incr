package incr

// list is a linked list datastructure that can be used
// as a FIFO queue or a stack depending on the context.
type list[K comparable, V any] struct {
	// head is the "first" element in the list
	head *listItem[K, V]
	// tail is the "last" element in the list
	tail *listItem[K, V]
	// lookup is a map between the key and the actual list item(s)
	lookup map[K]*listItem[K, V]
}

// Len returns the length of the list.
func (l *list[K, V]) Len() int {
	return len(l.lookup)
}

// Push appends a node to the end, or tail, of the list.
func (l *list[K, V]) Push(k K, v V) *listItem[K, V] {
	item := &listItem[K, V]{
		key:   k,
		value: v,
	}
	if l.lookup == nil {
		l.lookup = make(map[K]*listItem[K, V])
	}
	l.lookup[k] = item

	if l.head == nil {
		l.head = item
		l.tail = item
		return item
	}

	l.tail.next = item
	item.previous = l.tail
	item.next = nil
	l.tail = item
	return item
}

// Push appends a node to the front, or head, of the list.
func (l *list[K, V]) PushFront(k K, v V) *listItem[K, V] {
	item := &listItem[K, V]{
		key:   k,
		value: v,
	}
	if l.lookup == nil {
		l.lookup = make(map[K]*listItem[K, V])
	}
	l.lookup[k] = item

	if l.head == nil {
		l.head = item
		l.tail = item
		return item
	}

	l.head.previous = item
	item.next = l.head
	item.previous = nil
	l.head = item
	return item
}

// Pop removes the first, or head, element of the list.
func (l *list[K, V]) Pop() (k K, v V, ok bool) {
	if l.head == nil {
		return
	}

	k = l.head.key
	v = l.head.value
	ok = true
	delete(l.lookup, k)

	if l.head == l.tail {
		l.head = nil
		l.tail = nil
		return
	}

	next := l.head.next
	next.previous = nil
	l.head = next
	return
}

// PopBack removes the last, or tail, element of the list.
func (l *list[K, V]) PopBack() (k K, v V, ok bool) {
	if l.tail == nil {
		return
	}

	k = l.tail.key
	v = l.tail.value
	ok = true
	delete(l.lookup, k)

	if l.tail == l.head {
		l.head = nil
		l.tail = nil
		return
	}

	previous := l.tail.previous
	previous.next = nil
	l.tail = previous
	return
}

// PopAll removes all the elements, returning a slice.
func (l *list[K, V]) PopAll() (output []V) {
	ptr := l.head
	for ptr != nil {
		output = append(output, ptr.value)
		ptr = ptr.next
	}
	l.head = nil
	l.tail = nil
	clear(l.lookup)
	return
}

// Find returns the list item that matches a given key.
func (l *list[K, V]) Find(k K) *listItem[K, V] {
	n, ok := l.lookup[k]
	if ok {
		return n
	}
	return nil
}

// Remove removes an element with a given key from the list.
func (l *list[K, V]) Remove(k K) (ok bool) {
	if len(l.lookup) == 0 {
		return
	}

	var node *listItem[K, V]
	node, ok = l.lookup[k]
	if !ok {
		return
	}
	delete(l.lookup, k)

	// splice """out""" the node from the list
	if node.next != nil {
		// do things
	}
	if node.previous != nil {
		// do things
	}
	return
}

type listItem[K comparable, V any] struct {
	key K
	// value is the INode
	value V
	// next is the pointer towards the tail
	next *listItem[K, V]
	// previous is the pointer towards the head
	previous *listItem[K, V]
}
