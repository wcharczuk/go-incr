package incr

import (
	"testing"
)

func Test_Heap(t *testing.T) {
	h := NewHeap(
		MinLess[int],
		2, 1, 5,
	)
	h.Push(3)
	ItsEqual(t, 1, OkValue(h.Peek()))
	var values []int
	for h.Len() > 0 {
		values = append(values, OkValue(h.Pop()))
	}
	ItsEqual(t, []int{1, 2, 3, 5}, values)
}

func Test_Heap_FixAt(t *testing.T) {
	h := NewHeap(
		MaxLess[int],
		2, 1, 5, 4, 3,
	)
	h.Values[2] = 10
	h.FixAt(2)

	max, ok := h.Peek()
	ItsEqual(t, true, ok)
	ItsEqual(t, 10, max)
}

func Test_Heap_RemoveAt(t *testing.T) {
	h := NewHeap(
		MinLess[int],
		2, 1, 5,
	)

	indexOf := func(val int) (outputIndex int, ok bool) {
		for index, elem := range h.Values {
			if elem == val {
				outputIndex = index
				ok = true
				return
			}
		}
		return
	}

	removed, ok := h.RemoveAt(OkValue(indexOf(1)))
	ItsEqual(t, true, ok)
	ItsEqual(t, 1, removed)

	removed, ok = h.RemoveAt(OkValue(indexOf(2)))
	ItsEqual(t, true, ok)
	ItsEqual(t, 2, removed)

	removed, ok = h.RemoveAt(OkValue(indexOf(5)))
	ItsEqual(t, true, ok)
	ItsEqual(t, 5, removed)

	removed, ok = h.RemoveAt(OkValue(indexOf(-1)))
	ItsEqual(t, false, ok)
	ItsEqual(t, 0, removed)
}

func Test_Heap_PushPop(t *testing.T) {
	empty := NewHeap(MinLess[int])
	value, ok := empty.PushPop(10)
	ItsEqual(t, false, ok)
	ItsEqual(t, 0, value)

	value, ok = empty.PushPop(12)
	ItsEqual(t, true, ok)
	ItsEqual(t, 10, value)

	value, ok = empty.PushPop(13)
	ItsEqual(t, true, ok)
	ItsEqual(t, 12, value)

	h := NewHeap(
		MinLess[int],
		2, 1, 5, 4, 3,
	)

	value, ok = h.PushPop(10)
	ItsEqual(t, true, ok)
	ItsEqual(t, 1, value)

	value, ok = h.PushPop(15)
	ItsEqual(t, true, ok)
	ItsEqual(t, 2, value)

	value, ok = h.PushPop(12)
	ItsEqual(t, true, ok)
	ItsEqual(t, 3, value)

	value, ok = h.PushPop(14)
	ItsEqual(t, true, ok)
	ItsEqual(t, 4, value)

	value, ok = h.PushPop(11)
	ItsEqual(t, true, ok)
	ItsEqual(t, 5, value)

	value, ok = h.PushPop(1)
	ItsEqual(t, true, ok)
	ItsEqual(t, 10, value)

	value, ok = h.PushPop(1)
	ItsEqual(t, true, ok)
	ItsEqual(t, 1, value)
}
