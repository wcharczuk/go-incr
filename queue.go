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

// Len returns the number of elements in the queue.
//
// Use `Cap()` to return the length of the backing array itself.
func (q *queue[A]) Len() int {
	return q.size
}

// Cap returns the total capacity of the queue, including empty elements.
func (q *queue[A]) Cap() int {
	return len(q.array)
}

// Clear removes all elements from the Queue.
//
// It does _not_ reclaim any backing buffer length.
//
// To resize the backing buffer, use `Trim(size)`.
func (q *queue[A]) Clear() {
	clear(q.array)
	q.head = 0
	q.tail = 0
	q.size = 0
}

// Trim trims a queue to a given size.
func (q *queue[A]) Trim(size int) {
	q.setCapacity(size)
}

// Push adds an element to the "back" of the Queue.
func (q *queue[A]) Push(v A) {
	if len(q.array) == 0 {
		q.array = make([]A, queueDefaultCapacity)
	} else if q.size == len(q.array) {
		q.setCapacity(len(q.array) << 1)
	}
	q.array[q.tail] = v
	q.tail = (q.tail + 1) % len(q.array)
	q.size++
}

// Pop removes the first (oldest) element from the Queue.
func (q *queue[A]) Pop() (output A, ok bool) {
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

// Pop removes the last (newest) element from the Queue.
func (q *queue[A]) PopBack() (output A, ok bool) {
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

// Peek returns but does not remove the first element.
func (q *queue[A]) Peek() (output A, ok bool) {
	if q.size == 0 {
		return
	}
	output = q.array[q.head]
	ok = true
	return
}

// PeekBack returns but does not remove the last element.
func (q *queue[A]) PeekBack() (output A, ok bool) {
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

// Values collects the storage array into a copy array which is returned.
func (q *queue[A]) Values() (output []A) {
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

// Each calls the fn for each element in the buffer.
func (q *queue[A]) Each(fn func(A)) {
	q.EachUntil(func(v A) bool { fn(v); return true })
}

// Each calls the fn for each element in the buffer.
func (q *queue[A]) EachUntil(fn func(A) bool) {
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

// ReverseEach calls fn in reverse order (tail to head).
func (q *queue[A]) ReverseEach(fn func(A)) {
	q.ReverseEachUntil(func(v A) bool { fn(v); return true })
}

// ReverseEachUntil calls fn in reverse order (tail to head)
// with the function able to abort iteration by returning `false`.
func (q *queue[A]) ReverseEachUntil(fn func(A) bool) {
	if q.size == 0 {
		return
	}
	if q.head < q.tail {
		for cursor := q.tail - 1; cursor >= q.head; cursor-- {
			if !fn(q.array[cursor]) {
				return
			}
		}
	} else {
		for cursor := q.tail - 1; cursor >= 0; cursor-- {
			if !fn(q.array[cursor]) {
				return
			}
		}
		for cursor := len(q.array) - 1; cursor >= q.head; cursor-- {
			if !fn(q.array[cursor]) {
				return
			}
		}
	}
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
