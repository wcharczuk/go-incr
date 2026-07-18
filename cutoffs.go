package incr

import "context"

// CutoffAlways stops propagation past a node unconditionally.
//
// The node still recomputes, so anything it does as a side effect still happens; only its
// dependents are spared. That makes it a way to read a value without letting it drive a
// computation.
func CutoffAlways[A any](scope Scope, input Incr[A]) Incr[A] {
	return CutoffContext(scope, input, func(_ context.Context, _, _ A) (bool, error) {
		return true, nil
	})
}

// CutoffNever propagates every recomputation, which is the default behavior written
// explicitly.
//
// Worth naming so that a graph can say it means it, rather than leaving a reader to
// wonder whether a cutoff was forgotten.
func CutoffNever[A any](scope Scope, input Incr[A]) Incr[A] {
	return CutoffContext(scope, input, func(_ context.Context, _, _ A) (bool, error) {
		return false, nil
	})
}

// CutoffEqual stops propagation when a node's new value equals its old one.
//
// This is the cutoff most graphs want, and its absence is why a write of an unchanged
// value otherwise costs a full propagation. See the package documentation on cutoffs.
func CutoffEqual[A comparable](scope Scope, input Incr[A]) Incr[A] {
	return CutoffContext(scope, input, func(_ context.Context, previous, next A) (bool, error) {
		return previous == next, nil
	})
}

// CutoffEqualFunc is [CutoffEqual] for types with no == , taking the comparison.
func CutoffEqualFunc[A any](scope Scope, input Incr[A], equal func(a, b A) bool) Incr[A] {
	return CutoffContext(scope, input, func(_ context.Context, previous, next A) (bool, error) {
		return equal(previous, next), nil
	})
}
