package incr

// All collects many incrementals into a single incremental slice, in the order
// given.
//
// Producing the slice means reading every input, so a change costs O(inputs) by
// construction; this exists for the convenience rather than to avoid that. When
// the goal is an aggregate rather than the values themselves, prefer
// [UnorderedArrayFold] for O(1) updates or [ReduceBalanced] for O(log n).
//
// Returns nil for no inputs.
func All[A any](scope Scope, inputs ...Incr[A]) Incr[[]A] {
	if len(inputs) == 0 {
		return nil
	}
	return MapN(scope, func(values ...A) []A {
		// copied because MapN reuses its value slice between passes
		out := make([]A, len(values))
		copy(out, values)
		return out
	}, inputs...)
}
