package incr

func newSet[T comparable](values ...T) set[T] {
	output := make(set[T], len(values))
	for _, v := range values {
		output[v] = struct{}{}
	}
	return output
}

type set[T comparable] map[T]struct{}

func (s set[T]) has(t T) (ok bool) {
	_, ok = s[t]
	return
}

func (s set[T]) add(t T) {
	s[t] = struct{}{}
}

func (s set[T]) copy() set[T] {
	output := make(set[T], len(s))
	for k := range s {
		output[k] = struct{}{}
	}
	return output
}

func (s set[T]) keys() (out []T) {
	out = make([]T, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	return
}
