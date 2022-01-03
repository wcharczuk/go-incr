package incr

import "constraints"

// Max returns the maximum of a given set of values.
func Max[A constraints.Ordered](values ...A) (output A) {
	if len(values) == 0 {
		return
	}
	output = values[0]
	for _, v := range values[1:] {
		if v > output {
			output = v
		}
	}
	return
}
