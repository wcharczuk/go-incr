package incr

// list is a linked list datastructure that can be used
// as a FIFO queue or a stack depending on the context.
type list[K comparable, V any] struct {
	// head is the "first" element in the list
	head *listItem[K, V]
	// tail is the "last" element in the list
	tail *listItem[K, V]
	// len is the overall length of the list
	len int
}

// Push appends a node to the end, or tail, of the list.
func (l *list[K, V]) Push(k K, v V) *listItem[K, V] {
	item := &listItem[K, V]{
		key:   k,
		value: v,
	}

	l.len++
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

	l.len++
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

	l.len--

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

	l.len--
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
	l.len = 0
	return
}

// Find returns the list item that matches a given key.
func (l *list[K, V]) Find(k K) *listItem[K, V] {
	ptr := l.head
	for ptr != nil {
		if ptr.key == k {
			return ptr
		}
		ptr = ptr.next
	}
	return nil
}

// Remove removes an element from the list.
func (l *list[K, V]) Remove(i *listItem[K, V]) {
	l.len--

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

	if l.head == i && l.tail == i {
		l.head = nil
		l.tail = nil
		return
	}
	if l.head == i {
		l.head = i.next
		return
	}
	if l.tail == i {
		l.tail = i.previous
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

type listItem[K comparable, V any] struct {
	key K
	// value is the INode
	value V
	// next is the pointer towards the tail
	next *listItem[K, V]
	// previous is the pointer towards the head
	previous *listItem[K, V]
}
