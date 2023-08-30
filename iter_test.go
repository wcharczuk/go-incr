package incr

import (
	"testing"

	"github.com/wcharczuk/go-incr/testutil"
)

func Test_filter(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	fn := func(i int) bool { return i > 2 }
	output := filter(input, fn)
	testutil.ItsEqual(t, []int{3, 4, 5}, output)
}

func Test_filterRemoved(t *testing.T) {
	input := []int{1, 2, 3, 4, 5}
	fn := func(i int) bool { return i > 2 }
	included, excluded := filterRemoved(input, fn)
	testutil.ItsEqual(t, []int{3, 4, 5}, included)
	testutil.ItsEqual(t, []int{1, 2}, excluded)
}
