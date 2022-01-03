package incr

// Set is a generic set.
type Set[A comparable] map[A]struct{}

// Add adds a given element.
func (s *Set[A]) Add(v A) {
	(*s)[v] = struct{}{}
}

// Has returns if a given element exists.
func (s *Set[A]) Has(v A) bool {
	_, ok := (*s)[v]
	return ok
}
