package incr

import "testing"

func Test_Set(t *testing.T) {
	s := make(Set[int])
	itsEqual(t, false, s.Has(1))
	s.Add(1)
	itsEqual(t, true, s.Has(1))
}
