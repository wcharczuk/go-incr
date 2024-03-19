package slicei

import (
	"github.com/wcharczuk/go-incr"
)

// AccumulateSorted returns a new incremental that continually sorts new values into
// an output array using an online insertion sort.
func AccumulateSorted[A any](scope incr.Scope, from incr.Incr[A], f SortFunc[A]) incr.Incr[[]A] {
	return Accumulate(scope, from, func(values []A, newValue A) []A {
		return insertionSort(values, newValue, f)
	})
}

// SortFunc is a function that can be used to sort slices.
type SortFunc[A any] func(a, b A) int

// Asc returns a sorted comparer for sortable values in ascending order.
func Asc[A ~int | ~float64 | ~string](testValue, newValue A) int {
	if testValue == newValue {
		return 0
	}
	if testValue < newValue {
		return -1
	}
	return 1
}

// Desc returns a sorted comparer for sortable values in descending order.
func Desc[A ~int | ~float64 | ~string](testValue, newValue A) int {
	if testValue == newValue {
		return 0
	}
	if testValue < newValue {
		return 1
	}
	return -1
}

func insertionSort[A any](values []A, newValue A, f SortFunc[A]) (output []A) {
	if len(values) == 0 {
		output = []A{newValue}
		return
	}
	insertionIndex := searchForInsert(values, newValue, f)
	output = make([]A, len(values)+1)
	copy(output, values[:insertionIndex])
	output[insertionIndex] = newValue
	copy(output[insertionIndex+1:], values[insertionIndex:])
	return output
}

func searchForInsert[A any](values []A, newValue A, f SortFunc[A]) int {
	i, j := 0, len(values)
	for i < j {
		h := int(uint(i+j) >> 1)
		if f(values[h], newValue) < 0 {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}
