package slicei

import "github.com/wcharczuk/go-incr"

// Filter takes an input incremental and applies a predicate to it.
func Filter[A any](g incr.Scope, input incr.Incr[[]A], pred func(A) bool) incr.Incr[[]A] {
	return incr.Map(g, input, func(value []A) []A {
		var output []A
		for _, v := range value {
			if pred(v) {
				output = append(output, v)
			}
		}
		return output
	})
}
