package incrutil

import (
	"fmt"
	"testing"

	"github.com/wcharczuk/go-incr"
	"github.com/wcharczuk/go-incr/testutil"
)

func Test_Sorted_asc(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, 0)
	s := Sorted(g, v, Asc)
	os := incr.MustObserve(g, s)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0}, os.Value())

	v.Set(1)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0, 1}, os.Value())

	v.Set(3)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0, 1, 3}, os.Value())

	v.Set(2)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0, 1, 2, 3}, os.Value())
}

func Test_Sorted_desc(t *testing.T) {
	ctx := testContext()
	g := incr.New()

	v := incr.Var(g, 0)
	s := Sorted(g, v, Desc)
	os := incr.MustObserve(g, s)

	err := g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{0}, os.Value())

	v.Set(1)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{1, 0}, os.Value())

	v.Set(3)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{3, 1, 0}, os.Value())

	v.Set(2)

	err = g.Stabilize(ctx)
	testutil.NoError(t, err)
	testutil.Equal(t, []int{3, 2, 1, 0}, os.Value())
}

func Test_insertionSort(t *testing.T) {
	testCases := [...]struct {
		Values   []int
		NewValue int
		Expected []int
	}{
		{nil, 0, []int{0}},
		{nil, 1, []int{1}},
		{nil, 2, []int{2}},
		{[]int{1}, 0, []int{0, 1}},
		{[]int{1}, 1, []int{1, 1}},
		{[]int{1}, 2, []int{1, 2}},
		{[]int{2}, 1, []int{1, 2}},
		{[]int{2}, 2, []int{2, 2}},
		{[]int{2}, 3, []int{2, 3}},
		{[]int{1, 2}, 3, []int{1, 2, 3}},
		{[]int{1, 2}, 0, []int{0, 1, 2}},
		{[]int{1, 3}, 2, []int{1, 2, 3}},
		{[]int{1, 3, 4}, 2, []int{1, 2, 3, 4}},
		{[]int{1, 2, 4}, 3, []int{1, 2, 3, 4}},
		{[]int{2, 3, 4}, 1, []int{1, 2, 3, 4}},
		{[]int{1, 2, 3}, 4, []int{1, 2, 3, 4}},
	}

	for _, tc := range testCases {
		actual := insertionSort(tc.Values, tc.NewValue, func(searchVal, newValue int) bool {
			return searchVal > newValue
		})
		testutil.Equal(t, tc.Expected, actual, fmt.Sprintf("values=%#v new_value=%v", tc.Values, tc.NewValue))
	}
}
