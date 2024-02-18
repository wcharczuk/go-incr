package incr

const (
	queueDefaultCapacity = 4
)

// queue implements a fifo buffer datastructure using a ringbuffer.
//
// Typically you would see queue's implemented with a linked list, but in
// garbage collected languages there are advantages to keeping an array of items
// live to prevent excess allocations which makes a ringbuffer faster in practice.
type queue[A any] struct {
	array []A
	head  int
	tail  int
	size  int
}

// len returns the number of items in the queue.
func (q *queue[A]) len() int {
	return q.size
}

// cap is the full size of the ringbuffer's backing
// array (which is typically more in practice than
// the queue's actual length).
func (q *queue[A]) cap() int {
	return len(q.array)
}

// clear empties the queue, but does not reclaim any
// allocated space.
//
// to resize the queue to a smaller size, you can use
// the `setCapacity(...)` function.
func (q *queue[A]) clear() {
	clear(q.array)
	q.head = 0
	q.tail = 0
	q.size = 0
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
