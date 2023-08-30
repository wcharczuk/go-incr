package incr

// filter applies a predicate to a given slice
// returning just the elements that the predicate returned
// `true` for.
func filter[T any](values []T, fn func(T) bool) []T {
	var output []T
	for x := 0; x < len(values); x++ {
		val := values[x]
		if fn(val) {
			output = append(output, val)
		}
	}
	return output
}

// filterRemoved applies a predicate to a given slice
// returning both the elements for which the predicate passed and
// the elements for which the predicate failed.
func filterRemoved[T any](values []T, fn func(T) bool) (include, exclude []T) {
	for x := 0; x < len(values); x++ {
		val := values[x]
		if fn(val) {
			include = append(include, val)
		} else {
			exclude = append(exclude, val)
		}
	}
	return
}
