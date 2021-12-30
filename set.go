package incr

// Set is a generic implementation of a map
// that only holds keys.
type Set[A comparable] struct {
	m map[A]struct{}
}

// Has returns if a given key is present.
func (s *Set[A]) Has(v A) (ok bool) {
	if s.m == nil {
		s.m = make(map[A]struct{})
	}
	_, ok = s.m[v]
	return
}

// Set adds a key.
func (s *Set[A]) Set(v A) {
	if s.m == nil {
		s.m = make(map[A]struct{})
	}
	s.m[v] = struct{}{}
}

// Del removes a key.
func (s *Set[A]) Del(v A) {
	delete(s.m, v)
}
