package incr

import "testing"

func Test_diffSlice(t *testing.T) {
	var s []int

	val, last := diffSlice(0, s)
	ItsEqual(t, 0, last)
	ItsEqual(t, 0, len(val))

	s = []int{
		1, 2, 3,
	}

	val, last = diffSlice(last, s)
	ItsEqual(t, 3, last)
	ItsEqual(t, 3, len(val))
	ItsEqual(t, []int{1, 2, 3}, val)

	s = []int{
		1, 2, 3, 4, 5,
	}

	val, last = diffSlice(last, s)
	ItsEqual(t, 5, last)
	ItsEqual(t, 2, len(val))
	ItsEqual(t, []int{4, 5}, val)
}
