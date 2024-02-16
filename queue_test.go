package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func okValue[K any](v K, ok bool) K {
	return v
}

func ok[K any](v K, ok bool) bool {
	return ok
}

func Test_queue(t *testing.T) {
	buffer := new(queue[int])

	buffer.push(1)
	testutil.Equal(t, 1, buffer.array[0], "the 0 index should hold the head value of 1")
	testutil.Equal(t, 1, buffer.len())
	testutil.Equal(t, 1, okValue(buffer.peek()))
	testutil.Equal(t, 1, okValue(buffer.peekBack()))

	buffer.push(2)
	testutil.Equal(t, 2, buffer.len())
	testutil.Equal(t, 1, okValue(buffer.peek()))
	testutil.Equal(t, 2, okValue(buffer.peekBack()))

	buffer.push(3)
	testutil.Equal(t, 3, buffer.len())
	testutil.Equal(t, 1, okValue(buffer.peek()))
	testutil.Equal(t, 3, okValue(buffer.peekBack()))

	buffer.push(4)
	testutil.Equal(t, 4, buffer.len())
	testutil.Equal(t, 1, okValue(buffer.peek()))
	testutil.Equal(t, 4, okValue(buffer.peekBack()))

	buffer.push(5)
	testutil.Equal(t, 5, buffer.len())
	testutil.Equal(t, 1, okValue(buffer.peek()))
	testutil.Equal(t, 5, okValue(buffer.peekBack()))

	buffer.push(6)
	testutil.Equal(t, 6, buffer.len())
	testutil.Equal(t, 1, okValue(buffer.peek()))
	testutil.Equal(t, 6, okValue(buffer.peekBack()))

	buffer.push(7)
	testutil.Equal(t, 7, buffer.len())
	testutil.Equal(t, 1, okValue(buffer.peek()))
	testutil.Equal(t, 7, okValue(buffer.peekBack()))

	buffer.push(8)
	testutil.Equal(t, 8, buffer.len())
	testutil.Equal(t, 1, okValue(buffer.peek()))
	testutil.Equal(t, 8, okValue(buffer.peekBack()))

	value, ok := buffer.pop()
	testutil.Equal(t, 0, buffer.array[0], "we should zero elements that we pop")
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 1, value)
	testutil.Equal(t, 7, buffer.len())
	testutil.Equal(t, 2, okValue(buffer.peek()))
	testutil.Equal(t, 8, okValue(buffer.peekBack()))

	value, ok = buffer.pop()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 2, value)
	testutil.Equal(t, 6, buffer.len())
	testutil.Equal(t, 3, okValue(buffer.peek()))
	testutil.Equal(t, 8, okValue(buffer.peekBack()))

	value, ok = buffer.pop()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 3, value)
	testutil.Equal(t, 5, buffer.len())
	testutil.Equal(t, 4, okValue(buffer.peek()))
	testutil.Equal(t, 8, okValue(buffer.peekBack()))

	value, ok = buffer.pop()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, value)
	testutil.Equal(t, 4, buffer.len())
	testutil.Equal(t, 5, okValue(buffer.peek()))
	testutil.Equal(t, 8, okValue(buffer.peekBack()))

	value, ok = buffer.pop()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 5, value)
	testutil.Equal(t, 3, buffer.len())
	testutil.Equal(t, 6, okValue(buffer.peek()))
	testutil.Equal(t, 8, okValue(buffer.peekBack()))

	value, ok = buffer.pop()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 6, value)
	testutil.Equal(t, 2, buffer.len())
	testutil.Equal(t, 7, okValue(buffer.peek()))
	testutil.Equal(t, 8, okValue(buffer.peekBack()))

	value, ok = buffer.pop()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 7, value)
	testutil.Equal(t, 1, buffer.len())
	testutil.Equal(t, 8, okValue(buffer.peek()))
	testutil.Equal(t, 8, okValue(buffer.peekBack()))

	value, ok = buffer.pop()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 8, value)
	testutil.Equal(t, 0, buffer.len())
	testutil.Equal(t, 0, okValue(buffer.peek()))
	testutil.Equal(t, 0, okValue(buffer.peekBack()))
}

func Test_queue_push(t *testing.T) {
	q := new(queue[int])

	testutil.Equal(t, 0, len(q.array))
	testutil.Equal(t, 0, q.len())

	q.push(314)

	testutil.Equal(t, true, ok(q.peek()))
	testutil.Equal(t, 314, okValue(q.peek()))
	testutil.Equal(t, true, ok(q.peekBack()))
	testutil.Equal(t, 314, okValue(q.peekBack()))

	testutil.Equal(t, 0, q.head)
	testutil.Equal(t, 1, q.tail)
	testutil.Equal(t, 1, q.size)
	testutil.Equal(t, 4, len(q.array))
	testutil.Equal(t, 1, q.len())

	q.push(1)
	q.push(2)
	q.push(3)

	testutil.Equal(t, 0, q.head)
	testutil.Equal(t, 0, q.tail)
	testutil.Equal(t, 4, q.size)
	testutil.Equal(t, 4, len(q.array))

	testutil.Equal(t, true, ok(q.peek()))
	testutil.Equal(t, 314, okValue(q.peek()))
	testutil.Equal(t, true, ok(q.peekBack()))
	testutil.Equal(t, 3, okValue(q.peekBack()))

	testutil.Equal(t, 0, q.head)
	testutil.Equal(t, 0, q.tail)
	testutil.Equal(t, 4, q.size)
	testutil.Equal(t, 4, len(q.array))
}

func Test_queue_Pop_afterGrow(t *testing.T) {
	q := new(queue[int])
	for x := 0; x < queueDefaultCapacity+1; x++ {
		q.push(x)
	}
	testutil.Equal(t, queueDefaultCapacity+1, q.len())

	for x := 0; x < queueDefaultCapacity+1; x++ {
		value, ok := q.pop()
		testutil.Equal(t, true, ok)
		testutil.Equal(t, x, value)
	}

	value, ok := q.pop()
	testutil.Equal(t, false, ok)
	testutil.Equal(t, 0, value)
}

func Test_queue_popBack(t *testing.T) {
	buffer := new(queue[int])

	value, ok := buffer.popBack()
	testutil.Equal(t, false, ok)
	testutil.Equal(t, 0, value)

	buffer.push(1)
	buffer.push(2)
	buffer.push(3)

	value, ok = buffer.popBack()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 0, buffer.array[2], "PopBack should zero the tail before returning")
	testutil.Equal(t, 3, value)
	testutil.Equal(t, 2, buffer.len())

	value, ok = buffer.popBack()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 2, value)
	testutil.Equal(t, 1, buffer.len())

	value, ok = buffer.popBack()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 1, value)
	testutil.Equal(t, 0, buffer.len())

	value, ok = buffer.popBack()
	testutil.Equal(t, false, ok)
	testutil.Equal(t, 0, value)

	// do a popback with tail == 0
	q := new(queue[int])

	for x := 0; x < queueDefaultCapacity; x++ {
		q.push(x)
	}
	value, ok = q.popBack()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, queueDefaultCapacity-1, value)
}

func Test_queue_cap(t *testing.T) {
	buffer := new(queue[int])
	for x := 0; x <= queueDefaultCapacity+1; x++ {
		buffer.push(x)
	}

	testutil.Equal(t, 8, buffer.cap())
}

func Test_queue_clear(t *testing.T) {
	buffer := new(queue[int])

	// establish situation where head < tail
	// but there are valid elements
	for x := 0; x < 7; x++ {
		buffer.push(x)
	}
	testutil.Equal(t, 7, buffer.len())
	buffer.clear()

	testutil.Equal(t, 0, buffer.len())
	testutil.Equal(t, 0, okValue(buffer.peek()))
	testutil.Equal(t, 0, okValue(buffer.peekBack()))

	// get into situation where head > tail

	// grow the buffer to 8 cap
	for x := 0; x < 8; x++ {
		buffer.push(x)
	}
	// remove 5 elements
	for x := 0; x < 5; x++ {
		_, _ = buffer.pop()
	}
	// push another 4 elements
	// causing tail to wrap around
	for x := 0; x < 4; x++ {
		buffer.push(x)
	}
	testutil.Equal(t, 7, buffer.len())
	buffer.clear()

	testutil.Equal(t, 0, buffer.len())
	testutil.Equal(t, 0, okValue(buffer.peek()))
	testutil.Equal(t, 0, okValue(buffer.peekBack()))
}

func Test_queue_trim(t *testing.T) {
	buffer := new(queue[int])

	for x := 0; x < 7; x++ {
		buffer.push(x)
	}
	testutil.Equal(t, 7, buffer.len())

	buffer.trim(5)
	testutil.Equal(t, 5, buffer.len())

	value, ok := buffer.popBack()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 4, value)

	value, ok = buffer.pop()
	testutil.Equal(t, true, ok)
	testutil.Equal(t, 0, value)

	testutil.Equal(t, 3, buffer.len())
}

func Test_queue_values(t *testing.T) {
	q := new(queue[int])

	testutil.Equal(t, 0, len(q.values()))

	q.push(1)
	q.push(2)
	q.push(3)

	values := q.values()
	testutil.Equal(t, []int{1, 2, 3}, values)

	_, _ = q.pop()
	_, _ = q.pop()
	_, _ = q.pop()

	q.push(1)
	q.push(2)
	q.push(3)

	values = q.values()
	testutil.Equal(t, []int{1, 2, 3}, values)
}
