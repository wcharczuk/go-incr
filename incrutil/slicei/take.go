package slicei

import "github.com/wcharczuk/go-incr"

// First returns the first count elements from an incremental that is typed as an array.
func First[A any](scope incr.Scope, input incr.Incr[[]A], count int) incr.Incr[[]A] {
	return incr.Map(scope, input, func(values []A) []A {
		if len(values) < count {
			return values
		}
		output := make([]A, count)
		copy(output, values[:count])
		return output
	})
}

// BeforeSorted returns the elements before a point determined by a search function.
//
// The requirement for the input incremental is that it should already be sorted.
//
// The function should return true for elements you would like to skip, and the first element
// that returns false is the one that will mark the end of the list.
//
// An example sort function might be:
//
//	func(v int) bool {
//		return v >= 5
//	}
//
// For a list of `[]int{0,1,2,3,4,5,6,7,8,9}` this will return `[]int{0,1,2,3,4}`.
func BeforeSorted[A any](scope incr.Scope, input incr.Incr[[]A], fn func(A) bool) incr.Incr[[]A] {
	return incr.Map(scope, input, func(values []A) []A {
		if len(values) == 0 {
			return values
		}
		index := search(values, fn)
		output := make([]A, index)
		copy(output, values[:index])
		return output
	})
}

// Last returns the last count elements from an incremental that is typed as an array.
func Last[A any](scope incr.Scope, input incr.Incr[[]A], count int) incr.Incr[[]A] {
	return incr.Map(scope, input, func(values []A) []A {
		if len(values) < count {
			return values
		}
		output := make([]A, count)
		copy(output, values[count:])
		return output
	})
}

// AfterSorted returns the elements after a point determined by a search function.
//
// The requirement for the input incremental is that it should already be sorted.
//
// The function should return true for elements you would like to skip, and the first element
// that returns false is the one that will mark the beginning of the list.
//
// An example sort function might be:
//
//	func(v int) bool {
//		return v > 5
//	}
//
// For a list of `[]int{0,1,2,3,4,5,6,7,8,9}` this will return `[]int{6,7,8,9}`.
func AfterSorted[A any](scope incr.Scope, input incr.Incr[[]A], fn func(A) bool) incr.Incr[[]A] {
	return incr.Map(scope, input, func(values []A) []A {
		if len(values) == 0 {
			return values
		}
		index := search(values, fn)
		output := make([]A, len(values)-index)
		copy(output, values[index:])
		return output
	})
}

func search[A any](values []A, f func(searchValue A) bool) int {
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	i, j := 0, len(values)
	for i < j {
		h := int(uint(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if !f(values[h]) {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return i
}
