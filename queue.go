package incr

const (
	ringBufferMinimumGrow     = 4
	ringBufferGrowFactor      = 200
	ringBufferDefaultCapacity = 4
)

// Queue is a fifo buffer that is backed by a pre-allocated array, as opposed to a linked-list
// which would allocate a whole new struct for each element, which saves GC churn.
// Push can be O(n), Dequeue can be O(1).
type Queue[A any] struct {
	array []A
	head  int
	tail  int
	size  int
}

// Len returns the length of the ring buffer (as it is currently populated).
// Actual memory footprint may be different.
func (q *Queue[A]) Len() (len int) {
	return q.size
}

// Capacity returns the total size of the ring bufffer, including empty elements.
func (q *Queue[A]) Capacity() int {
	return len(q.array)
}

// Clear removes all objects from the RingBuffer.
func (q *Queue[A]) Clear() {
	if q.head < q.tail {
		arrayClear(q.array, q.head, q.size)
	} else {
		arrayClear(q.array, q.head, len(q.array)-q.head)
		arrayClear(q.array, 0, q.tail)
	}

	q.head = 0
	q.tail = 0
	q.size = 0
}

// Push adds an element to the "back" of the RingBuffer.
func (rb *Queue[A]) Push(object A) {
	if len(rb.array) == 0 {
		rb.array = make([]A, ringBufferDefaultCapacity)
	} else if rb.size == len(rb.array) {
		newCapacity := int(len(rb.array) * int(ringBufferGrowFactor/100))
		if newCapacity < (len(rb.array) + ringBufferMinimumGrow) {
			newCapacity = len(rb.array) + ringBufferMinimumGrow
		}
		rb.setCapacity(newCapacity)
	}
	rb.array[rb.tail] = object
	rb.tail = (rb.tail + 1) % len(rb.array)
	rb.size++
}

// Dequeue removes the first (oldest) element from the RingBuffer.
func (rb *Queue[A]) Pop() (output A, ok bool) {
	if rb.size == 0 {
		return
	}
	output = rb.array[rb.head]
	ok = true
	rb.head = (rb.head + 1) % len(rb.array)
	rb.size--
	return
}

// Peek returns but does not remove the first element.
func (rb *Queue[A]) Peek() (output A, ok bool) {
	if rb.size == 0 {
		return
	}
	output = rb.array[rb.head]
	ok = true
	return
}

// PeekBack returns but does not remove the last element.
func (rb *Queue[A]) PeekBack() (output A, ok bool) {
	if rb.size == 0 {
		return
	}
	if rb.tail == 0 {
		output = rb.array[len(rb.array)-1]
		ok = true
		return
	}
	output = rb.array[rb.tail-1]
	ok = true
	return
}

// Each calls the fn for each element in the buffer.
func (rb *Queue[A]) Each(fn func(A) error) (err error) {
	if rb.size == 0 {
		return
	}
	if rb.head < rb.tail {
		for cursor := rb.head; cursor < rb.tail; cursor++ {
			if err = fn(rb.array[cursor]); err != nil {
				return
			}
		}
	} else {
		for cursor := rb.head; cursor < len(rb.array); cursor++ {
			if err = fn(rb.array[cursor]); err != nil {
				return
			}
		}
		for cursor := 0; cursor < rb.tail; cursor++ {
			if err = fn(rb.array[cursor]); err != nil {
				return
			}
		}
	}
	return
}

// ReverseEach calls fn in reverse order (tail to head).
func (rb *Queue[A]) ReverseEach(fn func(A) error) (err error) {
	if rb.size == 0 {
		return
	}
	if rb.head < rb.tail {
		for cursor := rb.tail - 1; cursor >= rb.head; cursor-- {
			if err = fn(rb.array[cursor]); err != nil {
				return
			}
		}
	} else {
		for cursor := rb.tail; cursor > 0; cursor-- {
			if err = fn(rb.array[cursor]); err != nil {
				return
			}
		}
		for cursor := len(rb.array) - 1; cursor >= rb.head; cursor-- {
			if err = fn(rb.array[cursor]); err != nil {
				return
			}
		}
	}
	return
}

func (rb *Queue[A]) setCapacity(capacity int) {
	newArray := make([]A, capacity)
	if rb.size > 0 {
		if rb.head < rb.tail {
			arrayCopy(rb.array, rb.head, newArray, 0, rb.size)
		} else {
			arrayCopy(rb.array, rb.head, newArray, 0, len(rb.array)-rb.head)
			arrayCopy(rb.array, 0, newArray, len(rb.array)-rb.head, rb.tail)
		}
	}
	rb.array = newArray
	rb.head = 0
	if rb.size == capacity {
		rb.tail = 0
	} else {
		rb.tail = rb.size
	}
}

// trimExcess resizes the buffer to better fit the contents.
func (rb *Queue[A]) trimExcess() {
	threshold := float64(len(rb.array)) * 0.9
	if rb.size < int(threshold) {
		rb.setCapacity(rb.size)
	}
}

func arrayClear[A any](source []A, index, length int) {
	var zero A
	for x := 0; x < length; x++ {
		absoluteIndex := x + index
		source[absoluteIndex] = zero
	}
}

func arrayCopy[A any](source []A, sourceIndex int, destination []A, destinationIndex, length int) {
	for x := 0; x < length; x++ {
		from := sourceIndex + x
		to := destinationIndex + x
		destination[to] = source[from]
	}
}
