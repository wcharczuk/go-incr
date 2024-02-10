package incrutil

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_diffSliceByIndicesAdded(t *testing.T) {
	var s []int

	val, last := diffSliceByIndicesAdded(0, s)
	testutil.Equal(t, 0, last)
	testutil.Equal(t, 0, len(val))

	s = []int{
		1, 2, 3,
	}

	val, last = diffSliceByIndicesAdded(last, s)
	testutil.Equal(t, 3, last)
	testutil.Equal(t, 3, len(val))
	testutil.Equal(t, []int{1, 2, 3}, val)

	s = []int{
		1, 2, 3, 4, 5,
	}

	val, last = diffSliceByIndicesAdded(last, s)
	testutil.Equal(t, 5, last)
	testutil.Equal(t, 2, len(val))
	testutil.Equal(t, []int{4, 5}, val)
}
