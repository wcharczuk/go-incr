package incr

// list is a linked list structure that can be used
// as a ordered list as well as a constant time
// map using a similar technique to high throughput LRU queues.
type list[K comparable, V any] struct {
	// head is the "first" element in the list
	head *listItem[K, V]
	// tail is the "last" element in the list
	tail *listItem[K, V]
	// items is a map between the key and the actual list item(s)
	items map[K]*listItem[K, V]
}

func (l *list[K, V]) isEmpty() bool {
	return len(l.items) == 0
}

func (l *list[K, V]) len() int {
	if l == nil {
		return 0
	}
	return len(l.items)
}

func (l *list[K, V]) push(k K, v V) *listItem[K, V] {
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

func (l *list[K, V]) pushFront(k K, v V) *listItem[K, V] {
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

func (l *list[K, V]) pop() (k K, v V, ok bool) {
	if l.head == nil { // follows that items empty
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

func (l *list[K, V]) popBack() (k K, v V, ok bool) {
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

func (l *list[K, V]) popAll() (output []V) {
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

func (l *list[K, V]) each(fn func(K, V)) {
	ptr := l.head
	for ptr != nil {
		fn(ptr.key, ptr.value)
		ptr = ptr.next
	}
	return
}

func (l *list[K, V]) consume(fn func(K, V)) {
	ptr := l.head
	for ptr != nil {
		fn(ptr.key, ptr.value)
		ptr = ptr.next
	}
	l.head = nil
	l.tail = nil
	clear(l.items)
	return
}

func (l *list[K, V]) has(k K) (ok bool) {
	if l.items == nil {
		return
	}
	_, ok = l.items[k]
	return
}

func (l *list[K, V]) get(k K) (v V, ok bool) {
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

func (l *list[K, V]) remove(k K) (ok bool) {
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
	// next is the pointer towards the tail
	next *listItem[K, V]
	// previous is the pointer towards the head
	previous *listItem[K, V]
}
