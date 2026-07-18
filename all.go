package incr

// All collects many incrementals into a single incremental slice, in the order
// given.
//
// Producing the slice means reading every input, so a change costs O(inputs) by
// construction; this exists for the convenience rather than to avoid that. When
// the goal is an aggregate rather than the values themselves, prefer
// [UnorderedArrayFold] for O(1) updates or [ReduceBalanced] for O(log n).
//
// With no inputs the result is an empty slice. Unlike [ReduceBalanced], which has no
// value to report for an empty input, this returns a usable node so that a caller
// assembling inputs dynamically does not have to special-case the empty case.
func All[A any](scope Scope, inputs ...Incr[A]) Incr[[]A] {
	if len(inputs) == 0 {
		return Return(scope, []A{})
	}
	return MapN(scope, func(values ...A) []A {
		// copied because MapN reuses its value slice between passes
		out := make([]A, len(values))
		copy(out, values)
		return out
	}, inputs...)
}
