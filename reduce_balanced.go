package incr

// ReduceBalanced combines inputs pairwise through a balanced tree, so that one
// input changing costs O(log n) rather than O(n).
//
// This is the combinator to reach for when aggregating many inputs with an
// associative operation. [MapN] hands every value to its function on every pass,
// so a single input changing re-reads all of them; a balanced tree only recomputes
// the path from the changed leaf to the root. For a thousand inputs that is ten
// nodes rather than a thousand.
//
// reduce must be associative, because the grouping is chosen by the tree rather
// than by the caller. It need not be commutative: inputs are combined in the order
// given. It need not have an inverse either, which is what distinguishes this from
// [UnorderedArrayFold] -- min, max and concatenation work here and cannot be
// expressed there.
//
// Returns the single input unchanged for one input, and nil for none: with no identity
// element there is no value to report for an empty reduction. Callers assembling inputs
// dynamically should check for that, or use [All] whose empty case is well defined.
func ReduceBalanced[A any](scope Scope, reduce func(A, A) A, inputs ...Incr[A]) Incr[A] {
	if len(inputs) == 0 {
		return nil
	}
	// copied so that the caller's slice is not reordered underneath them
	level := make([]Incr[A], len(inputs))
	copy(level, inputs)
	for len(level) > 1 {
		next := level[:0:0]
		if cap(next) < (len(level)+1)/2 {
			next = make([]Incr[A], 0, (len(level)+1)/2)
		}
		var index int
		for ; index+1 < len(level); index += 2 {
			next = append(next, Map2(scope, level[index], level[index+1], reduce))
		}
		// an odd trailing input is carried up a level rather than combined with a
		// synthetic identity, which keeps the tree balanced without requiring the
		// caller to supply one.
		if index < len(level) {
			next = append(next, level[index])
		}
		level = next
	}
	return level[0]
}
