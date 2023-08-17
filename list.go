package incr

const (
	queueDefaultCapacity = 4
)

type listItem[K comparable, V any] struct {
	Key   K
	Value V
}

// List is a fifo buffer that is backed by a pre-allocated array, as opposed to a linked-list
// which would allocate a whole new struct for each element, which saves GC churn.
// Push can be O(n), Dequeue can be O(1).
type List[K comparable, V any] struct {
	array []listItem[K, V]
	head  int
	tail  int
	len   int
}

// Len returns the length of the queue (as it is currently populated).
//
// Actual memory footprint may be different, use `Cap()` to return total memory.
func (l *List[K, V]) Len() int {
	return l.len
}

// Cap returns the total capacity of the queue, including empty elements.
func (l *List[K, V]) Cap() int {
	return len(l.array)
}

// Clear removes all objects from the Queue.
func (l *List[K, V]) Clear() {
	if l.head < l.tail {
		arrayClear(l.array, l.head, l.len)
	} else {
		arrayClear(l.array, l.head, len(l.array)-l.head)
		arrayClear(l.array, 0, l.tail)
	}
	l.head = 0
	l.tail = 0
	l.len = 0
}

// PopAll collects the storage array into a copy array which is returned,
// zeroing any elements currently contained in the list.
func (l *List[K, V]) PopAll() (output []V) {
	if l.len == 0 {
		return
	}
	var zero listItem[K, V]
	output = make([]V, 0, l.len)
	if l.head < l.tail {
		for cursor := l.head; cursor < l.tail; cursor++ {
			output = append(output, l.array[cursor].Value)
			l.array[cursor] = zero
		}
	} else {
		for cursor := l.head; cursor < len(l.array); cursor++ {
			output = append(output, l.array[cursor].Value)
			l.array[cursor] = zero
		}
		for cursor := 0; cursor < l.tail; cursor++ {
			output = append(output, l.array[cursor].Value)
			l.array[cursor] = zero
		}
	}
	return
}

// Trim trims a queue to a given size.
func (l *List[K, V]) Trim(size int) {
	l.setCapacity(size)
}

// Push adds an element to the "back" of the Queue.
func (l *List[K, V]) Push(k K, v V) {
	if len(l.array) == 0 {
		l.array = make([]listItem[K, V], queueDefaultCapacity)
	} else if l.len == len(l.array) {
		l.setCapacity(len(l.array) << 1)
	}
	l.array[l.tail] = listItem[K, V]{k, v}
	l.tail = (l.tail + 1) % len(l.array)
	l.len++
}

// Pop removes the oldest (head) element from the Queue.
func (l *List[K, V]) Pop() (key K, val V, ok bool) {
	if l.len == 0 {
		return
	}

	key = l.array[l.head].Key
	val = l.array[l.head].Value
	ok = true

	var zero listItem[K, V]
	l.array[l.head] = zero
	l.head = (l.head + 1) % len(l.array)
	l.len--
	return
}

// PopBack removes the newest (tail) element from the Queue.
func (l *List[K, V]) PopBack() (key K, val V, ok bool) {
	if l.len == 0 {
		return
	}

	var zero listItem[K, V]
	if l.tail == 0 {
		key = l.array[len(l.array)-1].Key
		val = l.array[len(l.array)-1].Value
		l.array[len(l.array)-1] = zero
		l.tail = len(l.array) - 1
	} else {
		key = l.array[l.tail-1].Key
		val = l.array[l.tail-1].Value
		l.array[l.tail-1] = zero
		l.tail = l.tail - 1
	}
	ok = true
	l.len--
	return
}

// Head returns but does not remove the first element.
func (l *List[K, V]) Head() (key K, val V, ok bool) {
	if l.len == 0 {
		return
	}
	key = l.array[l.head].Key
	val = l.array[l.head].Value
	ok = true
	return
}

// Tail returns but does not remove the last element.
func (l *List[K, V]) Tail() (key K, val V, ok bool) {
	if l.len == 0 {
		return
	}
	if l.tail == 0 {
		key = l.array[len(l.array)-1].Key
		val = l.array[len(l.array)-1].Value
		ok = true
		return
	}
	key = l.array[l.tail-1].Key
	val = l.array[l.tail-1].Value
	ok = true
	return
}

// Each calls the fn for each element in the buffer.
func (l *List[K, V]) Each(fn func(K, V)) {
	if l.len == 0 {
		return
	}
	if l.head < l.tail {
		for cursor := l.head; cursor < l.tail; cursor++ {
			fn(l.array[cursor].Key, l.array[cursor].Value)
		}
	} else {
		for cursor := l.head; cursor < len(l.array); cursor++ {
			fn(l.array[cursor].Key, l.array[cursor].Value)
		}
		for cursor := 0; cursor < l.tail; cursor++ {
			fn(l.array[cursor].Key, l.array[cursor].Value)
		}
	}
}

// Values collects the storage array into a copy array which is returned.
func (l *List[K, V]) Values() (output []V) {
	if l.len == 0 {
		return
	}
	output = make([]V, 0, l.len)
	if l.head < l.tail {
		for cursor := l.head; cursor < l.tail; cursor++ {
			output = append(output, l.array[cursor].Value)
		}
	} else {
		for cursor := l.head; cursor < len(l.array); cursor++ {
			output = append(output, l.array[cursor].Value)
		}
		for cursor := 0; cursor < l.tail; cursor++ {
			output = append(output, l.array[cursor].Value)
		}
	}
	return
}

// setCapacity copies the queue into a new buffer
// with the given capacity.
//
// the new buffer will reset the head and tail
// indices such that head will be 0, and tail
// will be wherever the size index places it.
func (l *List[K, V]) setCapacity(capacity int) {
	newArray := make([]listItem[K, V], capacity)
	if l.len > 0 {
		if l.head < l.tail {
			copy(newArray, l.array[l.head:l.head+l.len])
		} else {
			copy(newArray, l.array[l.head:])
			copy(newArray[len(l.array)-l.head:], l.array[:l.tail])
		}
	}
	l.array = newArray
	l.head = 0
	if capacity < l.len {
		l.len = capacity
	}
	if l.len == capacity {
		l.tail = 0
	} else {
		l.tail = l.len
	}
}

func arrayClear[A any](source []A, index, length int) {
	var zero A
	for x := 0; x < length; x++ {
		absoluteIndex := x + index
		source[absoluteIndex] = zero
	}
}
