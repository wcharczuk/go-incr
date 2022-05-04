package incr

// Set is a simple hashset.
type Set[T comparable] map[T]struct{}

// Set adds an element to the set.
func (s Set[T]) Set(t T) {
	s[t] = struct{}{}
}

// Has returns if an element exists in a set.
func (s Set[T]) Has(t T) (ok bool) {
	_, ok = s[t]
	return
}

// Del removes an element from a set.
func (s Set[T]) Del(t T) {
	delete(s, t)
}
