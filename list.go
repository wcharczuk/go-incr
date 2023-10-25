package incr

import "sync"

// list is a linked list structure that can be used
// as a ordered list as well as a constant time
// map using a similar technique to high throughput LRU queues.
type list[K comparable, V any] struct {
	mu sync.Mutex
	// head is the "first" element in the list
	head *listItem[K, V]
	// tail is the "last" element in the list
	tail *listItem[K, V]
	// items is a map between the key and the actual list item(s)
	items map[K]*listItem[K, V]
}

// IsEmpty returns if the list has no items.
func (l *list[K, V]) IsEmpty() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.isEmptyUnsafe()
}

func (l *list[K, V]) isEmptyUnsafe() bool {
	return len(l.items) == 0
}

// Len returns the length of the list.
func (l *list[K, V]) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lenUnsafe()
}

func (l *list[K, V]) lenUnsafe() int {
	return len(l.items)
}

// Push appends a node to the end, or tail, of the list.
func (l *list[K, V]) Push(k K, v V) *listItem[K, V] {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pushUnsafe(k, v)
}

func (l *list[K, V]) pushUnsafe(k K, v V) *listItem[K, V] {
	item := &listItem[K, V]{
		key:   k,
		value: v,
	}
	if l.items == nil {
		l.items = make(map[K]*listItem[K, V])
	}
	l.items[k] = item

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

// PushFront appends a node to the front, or head, of the list.
func (l *list[K, V]) PushFront(k K, v V) *listItem[K, V] {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pushFrontUnsafe(k, v)
}

func (l *list[K, V]) pushFrontUnsafe(k K, v V) *listItem[K, V] {
	item := &listItem[K, V]{
		key:   k,
		value: v,
	}
	if l.items == nil {
		l.items = make(map[K]*listItem[K, V])
	}
	l.items[k] = item

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

// Pop removes the  head element of the list and returns it.
func (l *list[K, V]) Pop() (k K, v V, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.popUnsafe()
}

func (l *list[K, V]) popUnsafe() (k K, v V, ok bool) {
	if l.head == nil { // follows that items is nil
		return
	}

	k = l.head.key
	v = l.head.value
	ok = true
	delete(l.items, k)

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

// PopBack removes the tail element of the list and returns it.
func (l *list[K, V]) PopBack() (k K, v V, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.popBackUnsafe()
}

func (l *list[K, V]) popBackUnsafe() (k K, v V, ok bool) {
	if l.tail == nil { // follows that items is nil
		return
	}
	k = l.tail.key
	v = l.tail.value
	ok = true
	delete(l.items, k)

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
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.popAllUnsafe()
}

func (l *list[K, V]) popAllUnsafe() (output []V) {
	ptr := l.head
	for ptr != nil {
		output = append(output, ptr.value)
		ptr = ptr.next
	}
	l.head = nil
	l.tail = nil
	clear(l.items)
	return
}

// Has returns if a given key is present in the list.
func (l *list[K, V]) Has(k K) (ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.hasUnsafe(k)
}

func (l *list[K, V]) hasUnsafe(k K) (ok bool) {
	if l.items == nil {
		return
	}
	_, ok = l.items[k]
	return
}

// Get returns the list item value that matches a given key.
func (l *list[K, V]) Get(k K) (v V, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.getUnsafe(k)
}

func (l *list[K, V]) getUnsafe(k K) (v V, ok bool) {
	if l.items == nil {
		return
	}
	var n *listItem[K, V]
	n, ok = l.items[k]
	if !ok {
		return
	}
	v = n.value
	ok = true
	return
}

// Remove removes an element with a given key from the list.
func (l *list[K, V]) Remove(k K) (ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.removeUnsafe(k)
}

func (l *list[K, V]) removeUnsafe(k K) (ok bool) {
	if len(l.items) == 0 {
		return
	}

	var node *listItem[K, V]
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

// removeHeadItem removes the head pointer.
func (l *list[K, V]) removeHeadItem() {
	// if we have a single element,
	// we will need to change the tail
	// pointer as well
	if l.head == l.tail {
		l.head = nil
		l.tail = nil
		return
	}

	// remove from head
	towardsTail := l.head.next
	if towardsTail != nil {
		towardsTail.previous = nil
	}
	l.head = towardsTail
}

func (l *list[K, V]) removeLinkedItem(item *listItem[K, V]) {
	towardsHead := item.previous
	towardsTail := item.next
	if towardsHead != nil {
		towardsHead.next = towardsTail
	}
	if towardsTail != nil {
		towardsTail.previous = towardsHead
	}
	if l.tail == item {
		l.tail = item.previous
		if l.tail != nil {
			// it's the tail!
			l.tail.next = nil
		}
	}
}

type listItem[K comparable, V any] struct {
	// key is a unique identifier
	key K
	// value is the INode
	value V
	// height is used for moving node(s)
	height int
	// next is the pointer towards the tail
	next *listItem[K, V]
	// previous is the pointer towards the head
	previous *listItem[K, V]
}
