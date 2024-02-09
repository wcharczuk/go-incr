package incr

const (
	queueDefaultCapacity = 4
)

// queue is a fifo (first-in, first-out) buffer implementation.
//
// It is is backed by a pre-allocated array, which saves GC churn because the memory used
// to hold elements is not released unless the queue is trimmed.
//
// This stands in opposition to how queues are typically are implemented, which is as a linked list.
//
// As a result, `Push` can be O(n) if the backing array needs to be embiggened, though this should be relatively rare
// in pracitce if you're keeping a fixed queue size.
//
// Pop is generally O(1) because it just moves pointers around and nil's out elements.
type queue[A any] struct {
	array []A
	head  int
	tail  int
	size  int
}

func (q *queue[A]) len() int {
	return q.size
}

func (q *queue[A]) cap() int {
	return len(q.array)
}

func (q *queue[A]) clear() {
	clear(q.array)
	q.head = 0
	q.tail = 0
	q.size = 0
}

func (q *queue[A]) trim(size int) {
	q.setCapacity(size)
}

func (q *queue[A]) push(v A) {
	if len(q.array) == 0 {
		q.array = make([]A, queueDefaultCapacity)
	} else if q.size == len(q.array) {
		q.setCapacity(len(q.array) << 1)
	}
	q.array[q.tail] = v
	q.tail = (q.tail + 1) % len(q.array)
	q.size++
}

func (q *queue[A]) pop() (output A, ok bool) {
	if q.size == 0 {
		return
	}
	var zero A
	output = q.array[q.head]
	q.array[q.head] = zero
	ok = true
	q.head = (q.head + 1) % len(q.array)
	q.size--
	return
}

func (q *queue[A]) popBack() (output A, ok bool) {
	if q.size == 0 {
		return
	}

	var zero A
	if q.tail == 0 {
		output = q.array[len(q.array)-1]
		q.array[len(q.array)-1] = zero
		q.tail = len(q.array) - 1
	} else {
		output = q.array[q.tail-1]
		q.array[q.tail-1] = zero
		q.tail = q.tail - 1
	}
	ok = true
	q.size--
	return
}

func (q *queue[A]) peek() (output A, ok bool) {
	if q.size == 0 {
		return
	}
	output = q.array[q.head]
	ok = true
	return
}

func (q *queue[A]) peekBack() (output A, ok bool) {
	if q.size == 0 {
		return
	}
	if q.tail == 0 {
		output = q.array[len(q.array)-1]
		ok = true
		return
	}
	output = q.array[q.tail-1]
	ok = true
	return
}

func (q *queue[A]) values() (output []A) {
	if q.size == 0 {
		return
	}
	output = make([]A, 0, q.size)
	if q.head < q.tail {
		for cursor := q.head; cursor < q.tail; cursor++ {
			output = append(output, q.array[cursor])
		}
	} else {
		for cursor := q.head; cursor < len(q.array); cursor++ {
			output = append(output, q.array[cursor])
		}
		for cursor := 0; cursor < q.tail; cursor++ {
			output = append(output, q.array[cursor])
		}
	}
	return
}

func (q *queue[A]) each(fn func(A)) {
	q.eachUntil(func(v A) bool { fn(v); return true })
}

func (q *queue[A]) eachUntil(fn func(A) bool) {
	if q.size == 0 {
		return
	}
	if q.head < q.tail {
		for cursor := q.head; cursor < q.tail; cursor++ {
			if !fn(q.array[cursor]) {
				return
			}
		}
	} else {
		for cursor := q.head; cursor < len(q.array); cursor++ {
			if !fn(q.array[cursor]) {
				return
			}
		}
		for cursor := 0; cursor < q.tail; cursor++ {
			if !fn(q.array[cursor]) {
				return
			}
		}
	}
}

// filter removes items that don't pass a given predicate.
//
// filter scans the full array and collects passing items into
// a new array, that becomes the new backing store for the ring buffer.
//
// it is fairly inefficient, so generally you'll want to call this
// (1) time with a lot of logic baked into the predicate, versus many
// times for individual items you want to remove (i.e. use
// a multi-item lookup in the predicate if you can!)
func (q *queue[A]) filter(fn func(A) bool) {
	if q.size == 0 {
		return
	}

	filtered := make([]A, len(q.array))
	var filteredIndex int
	if q.head < q.tail {
		for cursor := q.head; cursor < q.tail; cursor++ {
			if fn(q.array[cursor]) {
				filtered[filteredIndex] = q.array[cursor]
				filteredIndex++
			}
		}
	} else {
		for cursor := q.head; cursor < len(q.array); cursor++ {
			if fn(q.array[cursor]) {
				filtered[filteredIndex] = q.array[cursor]
				filteredIndex++
			}
		}
		for cursor := 0; cursor < q.tail; cursor++ {
			if fn(q.array[cursor]) {
				filtered[filteredIndex] = q.array[cursor]
				filteredIndex++
			}
		}
	}
	q.array = filtered
	q.head = 0
	q.tail = filteredIndex
	q.size = filteredIndex
	return
}

// setCapacity copies the queue into a new buffer
// with the given capacity.
//
// the new buffer will reset the head and tail
// indices such that head will be 0, and tail
// will be wherever the size index places it.
func (q *queue[A]) setCapacity(capacity int) {
	newArray := make([]A, capacity)
	if q.size > 0 {
		if q.head < q.tail {
			copy(newArray, q.array[q.head:q.head+q.size])
		} else {
			copy(newArray, q.array[q.head:])
			copy(newArray[len(q.array)-q.head:], q.array[:q.tail])
		}
	}
	q.array = newArray
	q.head = 0
	if capacity < q.size {
		q.size = capacity
	}
	if q.size == capacity {
		q.tail = 0
	} else {
		q.tail = q.size
	}
}

func arrayClear[A any](source []A, index, length int) {
	var zero A
	for x := 0; x < length; x++ {
		absoluteIndex := x + index
		source[absoluteIndex] = zero
	}
}
