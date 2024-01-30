package incr

import "sync"

// list is a linked list structure that can be used
// as a ordered list as well as a constant time
// map using a similar technique to high throughput LRU queues.
//
// The exported functions on the list type are interlocked, that is
// they're safe to use oncurrently, and the unexported ...Unsafe versions are
// designed to be used internally when the root lock is held by the caller
// or another synchronization mechanism is used.
//
// You shouldn't ever need to interact with a list directly
// but we could eventually expose this as a structure people
// can take advantage of for their own use cases.
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

// Len returns the length of the list.
func (l *list[K, V]) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.lenUnsafe()
}

// Clear empties the list fully.
func (l *list[K, V]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.clearUnsafe()
}

// Push appends a node to the end, or tail, of the list.
func (l *list[K, V]) Push(k K, v V) *listItem[K, V] {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pushUnsafe(k, v)
}

// PushFront appends a node to the front, or head, of the list.
func (l *list[K, V]) PushFront(k K, v V) *listItem[K, V] {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.pushFrontUnsafe(k, v)
}

// Each iterates through each item in order.
func (l *list[K, V]) Each(fn func(V) error) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.eachUnsafe(fn)
}

// ConsumeEach iterates through each item in order
// leaving an empty list on completion.
func (l *list[K, V]) ConsumeEach(fn func(V)) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.consumeEachUnsafe(fn)
}

// Values returns the list contents as an array.
func (l *list[K, V]) Values() (out []V) {
	l.mu.Lock()
	defer l.mu.Unlock()
	out = l.valuesUnsafe()
	return
}

// Pop removes the  head element of the list and returns it.
func (l *list[K, V]) Pop() (k K, v V, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.popUnsafe()
}

// PopBack removes the tail element of the list and returns it.
func (l *list[K, V]) PopBack() (k K, v V, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.popBackUnsafe()
}

// PopAll removes all the elements, returning a slice.
func (l *list[K, V]) PopAll() (output []V) {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.popAllUnsafe()
}

// Has returns if a given key is held in the list
// and corresponds to an item.
//
// This call is interlocked with a mutex.
func (l *list[K, V]) Has(k K) (ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.hasUnsafe(k)
}

// Get returns the list item value that matches a given key.
//
// This call is interlocked with a mutex.
func (l *list[K, V]) Get(k K) (v V, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.getUnsafe(k)
}

// Remove removes an element with a given key from the list.
//
// This call is interlocked with a mutex.
func (l *list[K, V]) Remove(k K) (ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.removeUnsafe(k)
}

//
// internal & unsafe methods
//

func (l *list[K, V]) valuesUnsafe() (out []V) {
	out = make([]V, 0, len(l.items))
	cursor := l.head
	for cursor != nil {
		out = append(out, cursor.value)
		cursor = cursor.next
	}
	return
}

func (l *list[K, V]) pushUnsafe(k K, v V) *listItem[K, V] {
	if l.items == nil {
		l.items = make(map[K]*listItem[K, V])
	}
	if existing, ok := l.items[k]; ok {
		return existing
	}
	item := &listItem[K, V]{
		key:   k,
		value: v,
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

func (l *list[K, V]) isEmptyUnsafe() bool {
	return len(l.items) == 0
}

func (l *list[K, V]) lenUnsafe() int {
	return len(l.items)
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

func (l *list[K, V]) clearUnsafe() {
	l.head = nil
	l.tail = nil
	clear(l.items)
	return
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

func (l *list[K, V]) eachUnsafe(fn func(V) error) (err error) {
	cursor := l.head
	for cursor != nil {
		if err = fn(cursor.value); err != nil {
			return
		}
		cursor = cursor.next
	}
	return
}

func (l *list[K, V]) consumeEachUnsafe(fn func(V)) {
	ptr := l.head
	for ptr != nil {
		fn(ptr.value)
		ptr = ptr.next
	}
	l.head = nil
	l.tail = nil
	clear(l.items)
	return
}

func (l *list[K, V]) hasUnsafe(k K) (ok bool) {
	if l.items == nil {
		return
	}
	_, ok = l.items[k]
	return
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

// listItem is a item in a list.
type listItem[K comparable, V any] struct {
	// key is a unique identifier
	key K
	// value is the item for the list
	value V
	// next is the pointer towards the tail
	next *listItem[K, V]
	// previous is the pointer towards the head
	previous *listItem[K, V]
}
