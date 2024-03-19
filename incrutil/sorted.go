package incrutil

import (
	"github.com/wcharczuk/go-incr"
)

// Sorted returns a new incremental that continually sorts new values into
// an output array using an online insertion sort.
func Sorted[A any](scope incr.Scope, from incr.Incr[A], f func(searchValue, newValue A) bool) incr.Incr[[]A] {
	return Accumulate(scope, from, func(values []A, newValue A) []A {
		return insertionSort(values, newValue, f)
	})
}

// Asc returns a sorted comparer for sortable values in ascending order.
func Asc[A ~int | ~float64 | ~string](testValue, newValue A) bool {
	return testValue > newValue
}

// Desc returns a sorted comparer for sortable values in descending order.
func Desc[A ~int | ~float64 | ~string](testValue, newValue A) bool {
	return testValue <= newValue
}

func insertionSort[A any](values []A, newValue A, f func(searchValue, newValue A) bool) (output []A) {
	if len(values) == 0 {
		output = []A{newValue}
		return
	}
	insertionIndex := search(values, newValue, f)
	output = make([]A, len(values)+1)
	copy(output, values[:insertionIndex])
	output[insertionIndex] = newValue
	copy(output[insertionIndex+1:], values[insertionIndex:])
	return output
}

func search[A any](values []A, newValue A, f func(searchValue, newValue A) bool) int {
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	i, j := 0, len(values)
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if !f(values[h], newValue) {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return i
}
