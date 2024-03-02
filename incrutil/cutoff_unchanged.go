package incrutil

import "github.com/wcharczuk/go-incr"

// CutoffUnchanged returns a helper incremental form that cuts off computations
// if a comparable value hasn't changed.
func CutoffUnchanged[A comparable](scope incr.Scope, input incr.Incr[A]) incr.Incr[A] {
	return incr.Cutoff(scope, input, func(prev, curr A) bool {
		return prev == curr
	})
}
