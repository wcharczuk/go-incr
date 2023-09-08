package incr

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
