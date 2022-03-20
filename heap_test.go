package incr

import (
	"testing"
)

func Test_Heap(t *testing.T) {
	h := &Heap[int]{
		[]int{2, 1, 5},
		func(a, b int) bool { return a < b },
	}
	h.Init()
	h.Push(3)

	itsEqual(t, []int{1, 2, 5, 3}, h.Values)

	itsEqual(t, 1, OkValue(h.Peek()))
	var values []int
	for h.Len() > 0 {
		values = append(values, OkValue(h.Pop()))
	}
	itsEqual(t, []int{1, 2, 3, 5}, values)
}

func Test_Heap_Fix(t *testing.T) {
	h := &Heap[int]{
		[]int{2, 1, 5},
		func(a, b int) bool { return a < b },
	}
	h.Init()
	h.Push(3)

	itsEqual(t, []int{1, 2, 5, 3}, h.Values)

	h.Values[1] = 10
	h.Fix(1)
	itsEqual(t, []int{1, 3, 5, 10}, h.Values)
}

func Test_Heap_Remove(t *testing.T) {
	h := &Heap[int]{
		[]int{2, 1, 5},
		func(a, b int) bool { return a < b },
	}
	h.Init()
	h.Push(3)

	itsEqual(t, []int{1, 2, 5, 3}, h.Values)

	h.Values[1] = 10
	h.Remove(0)
	itsEqual(t, []int{1, 10, 5}, h.Values)
}

func Test_Heap_Peek(t *testing.T) {
	h := &Heap[int]{
		nil,
		func(a, b int) bool { return a < b },
	}
	v, ok := h.Peek()
	itsEqual(t, false, ok)
	itsEqual(t, 0, v)

	h.Push(123)

	v, ok = h.Peek()
	itsEqual(t, true, ok)
	itsEqual(t, 123, v)
}
