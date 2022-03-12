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

	itsEqual(t, 1, OkValue(h.Peek()))
	var values []int
	for h.Len() > 0 {
		values = append(values, OkValue(h.Pop()))
	}
	itsEqual(t, []int{1, 2, 3, 5}, values)
}
