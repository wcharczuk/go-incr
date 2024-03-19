package slicei

import (
	"slices"

	"github.com/wcharczuk/go-incr"
)

// Sort sorts a given input incremental according to the sort func.
func Sort[A any](scope incr.Scope, input incr.Incr[[]A], fn SortFunc[A]) incr.Incr[[]A] {
	return incr.Map(scope, input, func(values []A) []A {
		output := make([]A, len(values))
		copy(output, values)
		slices.SortFunc(output, fn)
		return output
	})
}
